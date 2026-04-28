package auth

// APIKey provides authentication using a static API key.
type APIKey struct {
	key string
}

// NewAPIKey creates a new APIKey authenticator.
func NewAPIKey(key string) (a *APIKey) {
	a = &APIKey{key: key}
	return
}

// Login is a no-op for API keys (they don't expire).
func (p *APIKey) Login() (err error) {
	return
}

// Header returns the Authorization header value.
func (p *APIKey) Header() (header string) {
	header = "Bearer " + p.key
	return
}
