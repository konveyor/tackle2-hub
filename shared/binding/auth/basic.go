package auth

import (
	"encoding/base64"
	"net/http"
)

// Basic provides HTTP Basic authentication.
type Basic struct {
	userid   string
	password string
}

// NewBasic creates a new Basic authenticator.
func NewBasic(userid, password string) (a *Basic) {
	a = &Basic{
		userid:   userid,
		password: password,
	}
	return
}

// Login is a no-op for Basic auth (credentials don't expire).
func (p *Basic) Login() (err error) {
	return
}

// Header returns the Authorization header value.
func (p *Basic) Header() (header string) {
	credentials := p.userid + ":" + p.password
	encoded := base64.URLEncoding.EncodeToString([]byte(credentials))
	header = "Basic " + encoded
	return
}

// SetTransport sets the HTTP transport for auth operations.
func (p *Basic) SetTransport(tp *http.Transport) {
	// No-op - Basic auth doesn't make HTTP calls.
}
