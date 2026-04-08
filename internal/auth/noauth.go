package auth

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// NewNoAuth return a provider.
func NewNoAuth(builtin *Builtin) *NoAuth {
	return &NoAuth{
		Builtin: builtin,
	}
}

// NoAuth provider always permits access.
type NoAuth struct {
	*Builtin
}

// Authenticate the token
func (r *NoAuth) Authenticate(request *Request) (jwToken *jwt.Token, err error) {
	return
}

// Scopes decodes a list of scopes from the token.
// For the NoAuth provider, this just returns a single
// wildcard scope matching everything.
func (r *NoAuth) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	scopes = append(scopes, &BaseScope{"*", "*"})
	return
}

// User mocks username for NoAuth
func (r *NoAuth) User(jwToken *jwt.Token) (name string) {
	name = "admin.noauth"
	return
}

// Handler returns an OIDC request handler.
func (r *NoAuth) Handler() (h http.Handler) {
	h = r.Builtin.Handler()
	return
}
