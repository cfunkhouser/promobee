package egobee

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	// ErrInvalidDuration is returned from UnmarshalJSON when the JSON does not
	// represent a Duration.
	ErrInvalidDuration = errors.New("invalid duration")

	hasUnitRx = regexp.MustCompile("[a-zA-Z]+")

	// now overrideable for testing.
	now = time.Now
)

// Scope of a token.
type Scope string

// Possible Scopes.
// See https://www.ecobee.com/home/developer/api/documentation/v1/auth/auth-intro.shtml
var (
	ScopeSmartRead  Scope = "smartRead"
	ScopeSmartWrite Scope = "smartWrite"
	ScopeEMSWrite   Scope = "ems"
)

// PinAuthenticationChallenge is the initial response from the Ecobee API for
// pin-based application authentication.
type PinAuthenticationChallenge struct {
	Pin               string `json:"ecobeePin"`
	AuthorizationCode string `json:"code"`
	Scope             Scope  `json:"scope"`
	// expires_in and interval are ignored for now.
}

// TokenDuration wraps time.Duration to add JSON (un)marshalling
type TokenDuration struct {
	time.Duration
}

// MarshalJSON returns JSON representation of Duration.
func (d TokenDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// UnmarshalJSON returns a Duration from JSON representation. Since the ecobee
// API returns durations in Seconds, values will be treated as such.
func (d *TokenDuration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Second * time.Duration(value)
	case string:
		if !hasUnitRx.Match([]byte(value)) {
			value = value + "s"
		}
		dv, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		d.Duration = dv
	default:
		return ErrInvalidDuration
	}
	return nil
}

// AuthorizationError returned by ecobee.
type AuthorizationError string

// Possible API Errors
var (
	AuthorizationErrorAccessDenied         AuthorizationError = "access_denied"
	AuthorizationErrorInvalidRequest       AuthorizationError = "invalid_request"
	AuthorizationErrorInvalidClient        AuthorizationError = "invalid_client"
	AuthorizationErrorInvalidGrant         AuthorizationError = "invalid_grant"
	AuthorizationErrorUnauthorizeClient    AuthorizationError = "unauthorized_client"
	AuthorizationErrorUnsupportedGrantType AuthorizationError = "unsupported_grant_type"
	AuthorizationErrorInvalidScope         AuthorizationError = "invalid_scope"
	AuthorizationErrorNotSupported         AuthorizationError = "not_supported"
	AuthorizationErrorAccountLocked        AuthorizationError = "account_locked"
	AuthorizationErrorAccountDisabled      AuthorizationError = "account_disabled"
	AuthorizationErrorAuthorizationPending AuthorizationError = "authorization_pending"
	AuthorizationErrorAuthorizationExpired AuthorizationError = "authorization_expired"
	AuthorizationErrorSlowDown             AuthorizationError = "slow_down"
)

// AuthorizationErrorResponse as returned from the ecobee API.
type AuthorizationErrorResponse struct {
	Error       AuthorizationError `json:"error"`
	Description string             `json:"error_description"`
	URI         string             `json:"error_uri"`
}

// Parse a response payload into the receiving AuthorizationErrorResponse. This will
// naturally fail if the payload is not an AuthorizationErrorResponse.
func (r *AuthorizationErrorResponse) Parse(payload []byte) error {
	if err := json.Unmarshal(payload, r); err != nil {
		return err
	}
	return nil
}

// ParseString behaves the same as Parse, but on a string.
func (r *AuthorizationErrorResponse) ParseString(payload string) error {
	return r.Parse([]byte(payload))
}

// Populate behaves the same as Parse, but reads the content from an io.Reader.
func (r *AuthorizationErrorResponse) Populate(reader io.Reader) error {
	d := json.NewDecoder(reader)
	return d.Decode(r)
}

// TokenRefreshResponse is returned by the ecobee API on toke refresh.
// See https://www.ecobee.com/home/developer/api/documentation/v1/auth/token-refresh.shtml
type TokenRefreshResponse struct {
	AccessToken  string        `json:"access_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    TokenDuration `json:"expires_in"`
	RefreshToken string        `json:"refresh_token"`
	Scope        Scope         `json:"scope"`
}

// Parse a response payload into the receiving TokenRefreshResponse. This will
// naturally fail if the payload is not an TokenRefreshResponse.
func (r *TokenRefreshResponse) Parse(payload []byte) error {
	if err := json.Unmarshal(payload, r); err != nil {
		return err
	}
	return nil
}

// ParseString behaves the same as Parse, but on a string.
func (r *TokenRefreshResponse) ParseString(payload string) error {
	return r.Parse([]byte(payload))
}

// Populate behaves the same as Parse, but reads the content from an io.Reader.
func (r *TokenRefreshResponse) Populate(reader io.Reader) error {
	d := json.NewDecoder(reader)
	return d.Decode(r)
}

// TokenStorer for ecobee Access and Refresh tokens.
type TokenStorer interface {
	// AccessToken gets the access token from the store.
	AccessToken() string

	// RefreshToken gets the refresh token from the store.
	RefreshToken() string

	// ValidFor reports how much longer the access token is valid.
	ValidFor() time.Duration

	// Update the TokenStorer with the contents of the response. This mutates the
	// access and refresh tokens.
	Update(*TokenRefreshResponse) error
}

// memoryStore implements tokenStore backed only by memory.
type memoryStore struct {
	mu           sync.RWMutex // protects the following members
	accessToken  string
	refreshToken string
	validUntil   time.Time
}

func (s *memoryStore) AccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken
}

func (s *memoryStore) RefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken
}

func (s *memoryStore) ValidFor() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.validUntil.Sub(now())
}

func (s *memoryStore) Update(r *TokenRefreshResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = r.AccessToken
	s.refreshToken = r.RefreshToken
	s.validUntil = generateValidUntil(r)

	return nil
}

// NewMemoryTokenStore is a TokenStorer with no persistence.
func NewMemoryTokenStore(r *TokenRefreshResponse) TokenStorer {
	s := &memoryStore{}
	s.Update(r)
	return s
}

const persistentStorePermissions = 0640

// persistentStoreData stores the data in memory matching the data stored to disk
type persistentStoreData struct {
	AccessTokenData  string    `json:"accessToken"`
	RefreshTokenData string    `json:"refreshToken"`
	ValidUntilData   time.Time `json:"validUntil"`
}

// persistentStore implements tokenStore backed by disk.
type persistentStore struct {
	mu   sync.RWMutex // protects the following members
	path string       // path to store file
	persistentStoreData
}

func (s *persistentStore) AccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AccessTokenData
}

func (s *persistentStore) RefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RefreshTokenData
}

func (s *persistentStore) ValidFor() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ValidUntilData.Sub(now())
}

func (s *persistentStore) Update(r *TokenRefreshResponse) error {
	// Update in-memory data
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AccessTokenData = r.AccessToken
	s.RefreshTokenData = r.RefreshToken
	s.ValidUntilData = generateValidUntil(r)

	f, err := os.OpenFile(s.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, persistentStorePermissions)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write token data to file to be accessed later
	return json.NewEncoder(f).Encode(&s.persistentStoreData)
}

// load the data from local file into memory.
func (s *persistentStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&s.persistentStoreData)
}

// NewPersistentTokenStore is a TokenStorer with persistence to disk
func NewPersistentTokenStore(r *TokenRefreshResponse, path string) (TokenStorer, error) {
	s := &persistentStore{
		path: path,
	}
	// update persistent storage tokenstore
	if err := s.Update(r); err != nil {
		return nil, err
	}

	return s, nil
}

// NewPersistentTokenFromDisk returns a TokenStorer based on disk location
func NewPersistentTokenFromDisk(path string) (TokenStorer, error) {
	s := &persistentStore{
		path: path,
	}
	return s, s.load()
}

// generateValidUntil returns the time the token expires with an added buffer
func generateValidUntil(r *TokenRefreshResponse) time.Time {
	// Subtract a few seconds to allow for network and processing delays.
	return now().Add(r.ExpiresIn.Duration - (15 * time.Second))
}
