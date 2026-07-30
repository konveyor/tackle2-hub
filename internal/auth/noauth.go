package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// NewNoAuth returns a NoAuth provider that permits all access.
func NewNoAuth(builtin *Builtin) *NoAuth {
	return &NoAuth{
		Builtin: builtin,
	}
}

// NoAuth provider always permits access.
type NoAuth struct {
	*Builtin
}

// Authenticate authenticates the request (always succeeds).
func (r *NoAuth) Authenticate(request *Request) (jwToken *jwt.Token, err error) {
	return
}

// Scopes decodes a list of scopes from the token.
// For the NoAuth provider, this just returns a single
// wildcard scope matching everything.
func (r *NoAuth) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	scopes = append(scopes, Scope{"*", "*"})
	return
}

// User returns the login for NoAuth provider.
func (r *NoAuth) User(jwToken *jwt.Token) (name string) {
	name = "admin.noauth"
	return
}

// Subject returns the subject for NoAuth provider.
func (r *NoAuth) Subject(jwToken *jwt.Token) (subject string) {
	return
}
