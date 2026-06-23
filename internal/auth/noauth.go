package auth

import (
	"net/http"
	"time"

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

func (r *NoAuth) NewToken(subject string, lifespan time.Duration) (token Token, err error) {
	token = r.newToken(subject, lifespan)
	err = r.Builtin.db.Create(&token).Error
	if err != nil {
		return
	}
	r.cache.TokenSaved(&token)
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

// Handler returns an OIDC request handler.
func (r *NoAuth) Handler() (h http.Handler) {
	h = r.Builtin.Handler()
	return
}

// IdpHandler returns the external IdP handler.
func (r *NoAuth) IdpHandler() (h *FedIdpHandler) {
	h = r.Builtin.IdpHandler()
	return
}
