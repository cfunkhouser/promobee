// Package egobee encapsulates types and helper functions for interacting with
// the ecobee REST API in Go.
package egobee

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type apiBaseURL string

func (a apiBaseURL) URL(apiPath string) string {
	return string(a) + apiPath
}

const (
	ecobeeAPIHost apiBaseURL = "https://api.ecobee.com"

	// These API Paths are relative to the API Host above.
	thermostatSummaryURL = "/1/thermostatSummary"
	thermostatURL        = "/1/thermostat"
	tokenURL             = "/token"
)

type reauthResponse struct {
	Err  *AuthorizationErrorResponse
	Resp *TokenRefreshResponse
}

func (r *reauthResponse) ok() bool {
	if r == nil {
		return false
	}
	return r.Err == nil && r.Resp != nil
}

func (r *reauthResponse) err() error {
	if r.Err != nil && r.Err.Error != "" && r.Err.Description != "" {
		return fmt.Errorf("unable to re-authenticate: %v: %v", r.Err.Error, r.Err.Description)
	}
	return errors.New("unable to re-authenticate for unknown reasons")
}

func reauthResponseFromHTTPResponse(resp *http.Response) (*reauthResponse, error) {
	r := &reauthResponse{}
	if (resp.StatusCode / 100) != 2 {
		r.Err = &AuthorizationErrorResponse{}
		if err := r.Err.Populate(resp.Body); err != nil {
			return nil, err
		}
	} else {
		r.Resp = &TokenRefreshResponse{}
		if err := r.Resp.Populate(resp.Body); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// authorizingTransport is a RoundTripper which includes the Access token in the
// request headers as appropriate for accessing the ecobee API.
type authorizingTransport struct {
	auth      TokenStorer
	transport http.RoundTripper
	appID     string
	api       apiBaseURL
}

func (t *authorizingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.shouldReauth() {
		if err := t.reauth(); err != nil {
			return nil, err
		}
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", t.auth.AccessToken()))
	return t.transport.RoundTrip(req)
}

func (t *authorizingTransport) shouldReauth() bool {
	// TODO(cfunkhouser): make the timeout customizable.
	return (t.auth.ValidFor() < (time.Second * 15)) || (t.auth.AccessToken() == "")
}

func (t *authorizingTransport) sendReauth(url string) (*reauthResponse, error) {
	tokenURL := fmt.Sprintf("%v?grant_type=refresh_token&refresh_token=%v&client_id=%v", url, t.auth.RefreshToken(), t.appID)
	resp, err := http.Post(tokenURL, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return reauthResponseFromHTTPResponse(resp)
}

func (t *authorizingTransport) reauth() error {
	r, err := t.sendReauth(t.api.URL(tokenURL))
	if err != nil {
		return err
	}
	if !r.ok() {
		return r.err()
	}
	return t.auth.Update(r.Resp)
}

func simpleRequestID() string {
	return fmt.Sprintf("req@%v", time.Now().UnixNano())
}

// loggingTransport is a RoundTripper which wraps a RoundTripper and logs every
// HTTP request and response to a Logger.
type loggingTransport struct {
	l         *log.Logger
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	id := simpleRequestID()
	if rb, err := httputil.DumpRequest(req, true); err == nil {
		t.l.Printf("Outgoing Request %v:\n%+v\n<< END %v", id, string(rb), id)
	}
	r, err := t.transport.RoundTrip(req)
	if err != nil {
		t.l.Printf("Error %v: %v", id, err)
	} else if rb, err := httputil.DumpResponse(r, true); err == nil {
		t.l.Printf("Incoming Response to %v:\n%+v\n<< END resp %v", id, string(rb), id)
	}
	return r, err
}

// Options to New.
type Options struct {
	// APIHost for Ecobee API requests. Defaults to https://api.ecobee.com.
	APIHost string
	// Log all requests to LogTo if true.
	Log bool
	// LogTo gets all requests and responses to this Writer verbosely.
	LogTo io.Writer
}

func (o *Options) apiHost() apiBaseURL {
	if o == nil || o.APIHost == "" {
		return ecobeeAPIHost
	}
	return apiBaseURL(o.APIHost)
}

func (o *Options) log() (io.Writer, bool) {
	if o == nil {
		return nil, false
	}
	return o.LogTo, o.Log
}

// Client for the ecobee API.
type Client struct {
	api apiBaseURL
	http.Client
}

// New egobee client.
func New(appID string, ts TokenStorer, opts ...*Options) *Client {
	// Someday I'll realize it would have been easier just to have a second New
	// function instead of doing this variadic bullshit.
	var opt *Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	var trans http.RoundTripper = &authorizingTransport{
		auth:      ts,
		transport: http.DefaultTransport,
		appID:     appID,
		api:       opt.apiHost(),
	}
	if w, doLog := opt.log(); doLog {
		trans = &loggingTransport{
			l:         log.New(w, "", log.LstdFlags),
			transport: trans,
		}
	}
	return &Client{
		api: opt.apiHost(),
		Client: http.Client{
			Transport: trans,
		},
	}
}
