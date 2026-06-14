package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/api"
)

const (
	Scopes = "openid profile email offline_access"
)

// OIDC endpoints per:
// - RFC 8628
// - RFC 6749
const (
	DeviceAuthRoute = "/device_authorization"
	TokenRoute      = "/token"
)

// TokenError defines a token error.
type TokenError struct {
	Error string `json:"error"`
}

// Error codes per: RFC 8628
const (
	AuthPending = "authorization_pending"
	SlowDown    = "slow_down"
)

// NewOIDC creates a new OIDC bearer token authenticator.
func NewOIDC(issuerURL, clientId string) (h *OIDC) {
	h = &OIDC{
		issuerURL:  issuerURL,
		clientId:   clientId,
		httpClient: &http.Client{},
	}
	return
}

// OIDC provides OIDC authentication with automatic token refresh.
type OIDC struct {
	mutex        sync.Mutex
	issuerURL    string
	clientId     string
	httpClient   *http.Client
	accessToken  string
	refreshToken string
}

// Use sets the access token.
func (p *OIDC) Use(token string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.accessToken = token
}

// Token returns the access token.
func (p *OIDC) Token() (token string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	token = p.accessToken
	return
}

// Login refreshes the access token.
// DeviceLogin when refresh fails.
func (p *OIDC) Login() (err error) {
	var refreshToken string
	func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		refreshToken = p.refreshToken
	}()

	if refreshToken == "" {
		err = p.DeviceLogin()
		return
	}

	token := &Token{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientId:     p.clientId,
	}
	err = p.post(TokenRoute, token)
	if err != nil {
		err = p.DeviceLogin()
		return
	}
	func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		p.accessToken = token.AccessToken
		if token.RefreshToken != "" {
			p.refreshToken = token.RefreshToken
		}
	}()
	return
}

// Header returns the Authorization header value.
func (p *OIDC) Header() (header string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	header = "Bearer " + p.accessToken
	return
}

// SetTransport sets the http transport.
func (p *OIDC) SetTransport(tr *http.Transport) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.httpClient.Transport = tr
}

// DeviceLogin performs device authorization flow.
func (p *OIDC) DeviceLogin() (err error) {
	authReq := &DeviceAuth{
		ClientId: p.clientId,
		Scope:    Scopes,
	}
	err = p.post(DeviceAuthRoute, authReq)
	if err != nil {
		return
	}

	fmt.Printf("\nVisit: %s\n", authReq.VerificationURI)
	fmt.Printf("Enter code: %s\n\n", authReq.UserCode)

	err = p.getToken(authReq)
	return
}

// getToken polls for tokens.
func (p *OIDC) getToken(authReq *DeviceAuth) (err error) {
	seconds := max(authReq.Interval, 5)

	token := &Token{
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		DeviceCode: authReq.DeviceCode,
		ClientId:   p.clientId,
	}

	for {
		delay := time.Duration(seconds) * time.Second
		time.Sleep(delay)

		err = p.post(TokenRoute, token)
		if err == nil {
			p.mutex.Lock()
			p.accessToken = token.AccessToken
			if token.RefreshToken != "" {
				p.refreshToken = token.RefreshToken
			}
			p.mutex.Unlock()
			return
		}

		var restErr *api.RestError
		if errors.As(err, &restErr) {
			var tokenErr TokenError
			parseErr := json.Unmarshal([]byte(restErr.Body), &tokenErr)
			if parseErr == nil {
				switch tokenErr.Error {
				case AuthPending:
					continue
				case SlowDown:
					seconds += 5
					continue
				default:
					return
				}
			}
		}
		return
	}
}

// post makes a POST request to the OIDC endpoint.
func (p *OIDC) post(path string, object Form) (err error) {
	parsed, err := url.Parse(p.issuerURL)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	parsed.Path, err = url.JoinPath(parsed.Path, path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	request, err := http.NewRequest(
		http.MethodPost,
		parsed.String(),
		bytes.NewReader([]byte(object.Data().Encode())))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	request.Header.Set(api.ContentType, "application/x-www-form-urlencoded")
	request.Header.Set(api.Accept, api.MIMEJSON)

	response, err := p.httpClient.Do(request)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()

	status := response.StatusCode
	if status < 200 || status >= 300 {
		body, _ := io.ReadAll(response.Body)
		restErr := &api.RestError{
			Status: status,
			Method: http.MethodPost,
			Path:   path,
			Body:   string(body),
		}
		err = restErr
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// Form object.
type Form interface {
	Data() url.Values
}

// DeviceAuth authorization request/response per RFC 8628
type DeviceAuth struct {
	ClientId        string `json:"client_id"`
	Scope           string `json:"scope,omitempty"`
	DeviceCode      string `json:"device_code,omitempty"`
	UserCode        string `json:"user_code,omitempty"`
	VerificationURI string `json:"verification_uri,omitempty"`
	Interval        int    `json:"interval,omitempty"`
}

// Data returns formData.
func (d *DeviceAuth) Data() (v url.Values) {
	v = make(url.Values)
	v.Set("client_id", d.ClientId)
	if d.Scope != "" {
		v.Set("scope", d.Scope)
	}
	return
}

// Token request/response per RFC 6749
type Token struct {
	GrantType    string `json:"grant_type,omitempty"`
	DeviceCode   string `json:"device_code,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ClientId     string `json:"client_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
}

// Data returns formData.
func (d *Token) Data() (v url.Values) {
	v = make(url.Values)
	if d.GrantType != "" {
		v.Set("grant_type", d.GrantType)
	}
	if d.DeviceCode != "" {
		v.Set("device_code", d.DeviceCode)
	}
	if d.RefreshToken != "" {
		v.Set("refresh_token", d.RefreshToken)
	}
	if d.ClientId != "" {
		v.Set("client_id", d.ClientId)
	}
	return
}
