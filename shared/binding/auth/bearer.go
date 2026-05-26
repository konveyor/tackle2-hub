package auth

import (
	"net/http"
)

// NewBearer return a new static bearer token auth method.
func NewBearer(token string) (m *Bearer) {
	m = &Bearer{
		token: token,
	}
	return
}

// Bearer static bearer method.
type Bearer struct {
	token string
}

// Token returns the bearer token.
func (p *Bearer) Token() (token string) {
	token = p.token
	return
}

// Header returns the authorization header.
func (p *Bearer) Header() (h string) {
	h = "Bearer " + p.token
	return
}

// Login No-op.
func (p *Bearer) Login() (err error) {
	return
}

// SetTransport No-op.
func (p *Bearer) SetTransport(*http.Transport) {
}
