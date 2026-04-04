package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NoAuth provider always permits access.
type NoAuth struct {
	handler http.Handler
}

// UserKey returns a new key associated with a user.
func (r *NoAuth) UserKey(userId, password string, lifespan time.Duration) (key APIKey, err error) {
	return
}

// TaskKey returns a new key associated with a task.
func (r *NoAuth) TaskKey(taskId uint, lifespan time.Duration) (key APIKey, err error) {
	return
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
	h = r.handler
	return
}
