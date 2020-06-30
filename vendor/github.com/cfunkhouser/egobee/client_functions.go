package egobee

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	requestContentType = "application/json; charset=utf-8"
)

// These are simple stubs which allow for easy overrides in testing without a
// heavy mocks library.
var (
	httpNewRequest = http.NewRequest
	jsonMarshal    = json.Marshal

	errPagingUnimplemented = errors.New("multi-page responses unimplemented")

	// jsonDecode wraps the usual JSON decode workflow to make testing easier.
	jsonDecode = func(from io.Reader, to interface{}) error {
		if err := json.NewDecoder(from).Decode(to); err != nil {
			return fmt.Errorf("failed to decode JSON: %v", err)
		}
		return nil
	}
)

// page is used for paging in some APIs.
type page struct {
	Page       int `json:"page"`
	TotalPages int `json:"totalPages"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
}

// summarySelection wraps a Selection, and serializes to the format expected by
// the thermostatSummary API.
type summarySelection struct {
	Selection Selection `json:"selection,omitempty"`
}

func assembleSelectionURL(apiURL string, selection *Selection) (string, error) {
	ss := &summarySelection{
		Selection: *selection,
	}
	qb, err := jsonMarshal(ss)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`%v?json=%v`, apiURL, url.QueryEscape(string(qb))), nil
}

func assembleSelectionRequest(url string, s *Selection) (*http.Request, error) {
	u, err := assembleSelectionURL(url, s)
	if err != nil {
		return nil, err
	}
	r, err := httpNewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	r.Header.Add("Content-Type", requestContentType)
	return r, nil
}

// validateSelectionResponse validates that a http.Response resulting from a
// selection request is actually usable.
func validateSelectionResponse(res *http.Response) error {
	if (res.StatusCode / 100) != 2 {
		return fmt.Errorf("non-ok status response from API: %v %v", res.StatusCode, res.Status)
	}
	return nil
}

// ThermostatSummary retrieves a list of thermostat configuration and state
// revisions. This API request is a light-weight polling method which will only
// return the revision numbers for the significant portions of the thermostat
// data.
// See https://www.ecobee.com/home/developer/api/documentation/v1/operations/get-thermostat-summary.shtml
func (c *Client) ThermostatSummary() (*ThermostatSummary, error) {
	req, err := assembleSelectionRequest(c.api.URL(thermostatSummaryURL), &Selection{
		SelectionType: SelectionTypeRegistered,
		IncludeEquipmentStatus: true,
		IncludeAlerts: true,
	})
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to Do(): %v", err)
	}
	defer res.Body.Close()

	if err := validateSelectionResponse(res); err != nil {
		return nil, err
	}

	ts := &ThermostatSummary{}
	if err := jsonDecode(res.Body, ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// See https://www.ecobee.com/home/developer/api/documentation/v1/operations/get-thermostats.shtml
type pagedThermostatResponse struct {
	Page        page          `json:"page,omitempty"`
	Thermostats []*Thermostat `json:"thermostatList,omitempty"`
	Status      struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"status,omitempty"`
}

// Thermostats returns all Thermostat objects which match selection.
func (c *Client) Thermostats(selection *Selection) ([]*Thermostat, error) {
	req, err := assembleSelectionRequest(c.api.URL(thermostatURL), selection)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := validateSelectionResponse(res); err != nil {
		return nil, err
	}

	ptr := &pagedThermostatResponse{}

	if err := jsonDecode(res.Body, ptr); err != nil {
		return nil, err
	}

	if ptr.Page.Page != ptr.Page.TotalPages {
		// TODO(cfunkhouser): Handle paged responses.
		return nil, errPagingUnimplemented
	}
	return ptr.Thermostats, nil
}
