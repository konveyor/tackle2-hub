package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/konveyor/controller/pkg/logging"
	"strings"
)

var (
	// Log logger.
	Log = logging.WithName("auth")
	// Hub provider.
	Hub = &Builtin{}
	// Remote provider.
	Remote Provider
)

//
// Provider provides RBAC.
type Provider interface {
	// Authenticate authenticates and validates the token.
	Authenticate(token string) (jwToken *jwt.Token, err error)
	// Scopes extracts a list of scopes from the token.
	Scopes(jwToken *jwt.Token) []Scope
	// User extracts the user from token.
	User(jwToken *jwt.Token) (user string)
}

//
// NotAuthenticated is returned when a token cannot be authenticated.
type NotAuthenticated struct {
	Token string
}

func (e *NotAuthenticated) Error() (s string) {
	return fmt.Sprintf("Token %s not-valid.", e.Token)
}

func (e *NotAuthenticated) Is(err error) (matched bool) {
	_, matched = err.(*NotAuthenticated)
	return
}

//
// NotValid is returned when a token is not valid.
type NotValid struct {
	Token string
}

func (e *NotValid) Error() (s string) {
	return fmt.Sprintf("Token %s not-valid.", e.Token)
}

func (e *NotValid) Is(err error) (matched bool) {
	_, matched = err.(*NotValid)
	return
}

//
// Scope represents an authorization scope.
type Scope interface {
	// Allow determines whether the scope gives access to the resource with the method.
	Allow(resource string, method string) bool
	// HasModifier determines if the matched scopes include the
	// specified modifier.
	HasModifier(name string) bool
}

//
// BaseScope.
type BaseScope struct {
	resource string
	method   string
	mods     []string
}

//
// With parses a scope and populate fields.
// Format: <resource>:<method>:<modifier>:...
func (r *BaseScope) With(s string) {
	part := strings.Split(s, ":")
	n := len(part)
	if n > 0 {
		r.resource = part[0]
	}
	if n > 1 {
		r.method = part[1]
	}
	if n > 2 {
		r.mods = part[2:]
	}
	return
}

//
// Allow determines whether the scope gives access to the resource with the method.
func (r *BaseScope) Allow(resource string, method string) (b bool) {
	b = (r.resource == "*" || r.resource == resource) &&
		(r.method == "*" || r.method == method)
	return
}

//
//HasModifier determines if the matched scopes include the
// specified modifier.
func (r *BaseScope) HasModifier(name string) (b bool) {
	for _, m := range r.mods {
		if strings.ToLower(m) == strings.ToLower(name) {
			b = true
			break
		}
	}
	return
}
