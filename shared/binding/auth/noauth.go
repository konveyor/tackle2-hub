package auth

import "net/http"

type NoAuth struct {
}

// Login performs authentication and refreshes credentials.
func (m *NoAuth) Login() (err error) {
	return
}

// Header returns the Authorization header value.
func (m *NoAuth) Header() (header string) {
	return
}

// SetTransport sets the HTTP transport for auth operations.
func (m *NoAuth) SetTransport(tp *http.Transport) {
	// No-op - NoAuth doesn't make HTTP calls.
}
