package auth

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
