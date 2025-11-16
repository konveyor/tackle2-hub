package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jortel/go-utils/logr"
)

var (
	// Log logger.
	Log = logr.New("auth", Settings.Log.Auth)
	// Hub provider.
	Hub Provider
	// Remote provider.
	Remote Provider
)

func init() {
	Hub = &NoAuth{}
	Remote = &NoAuth{}
}

// Provider provides RBAC.
type Provider interface {
	// NewToken creates a signed token.
	NewToken(user string, scopes []string, claims jwt.MapClaims) (signed string, err error)
	// Authenticate authenticates and validates the token.
	Authenticate(r *Request) (jwToken *jwt.Token, err error)
	// Scopes extracts a list of scopes from the token.
	Scopes(jwToken *jwt.Token) []Scope
	// User extracts the user from token.
	User(jwToken *jwt.Token) (user string)
	// Login and obtain a token.
	Login(user, password string) (token Token, err error)
	// Refresh token.
	Refresh(refresh string) (token Token, err error)
}

type Token struct {
	Access  string
	Refresh string
	Expiry  int
}

// NotAuthenticated is returned when a token cannot be authenticated.
type NotAuthenticated struct {
	Token string
}

func (e *NotAuthenticated) Error() (s string) {
	return fmt.Sprintf("Token [%s] not-authenticated.", e.Token)
}

func (e *NotAuthenticated) Is(err error) (matched bool) {
	notAuth := &NotAuthenticated{}
	matched = errors.As(err, &notAuth)
	return
}

// NotValid is returned when a token is not valid.
type NotValid struct {
	Reason string
	Token  string
}

func (e *NotValid) Error() (s string) {
	return fmt.Sprintf("Token [%s] not-valid: %s", e.Token, e.Reason)
}

func (e *NotValid) Is(err error) (matched bool) {
	notValid := &NotValid{}
	matched = errors.As(err, &notValid)
	return
}

// Scope represents an authorization scope.
type Scope interface {
	// Match returns whether the scope is a match.
	Match(resource string, method string) bool
	//String representations of the scope.
	String() (s string)
}

// BaseScope provides base behavior.
type BaseScope struct {
	Resource string
	Method   string
}

// With parses a scope and populate fields.
// Format: <resource>:<method>
func (r *BaseScope) With(s string) {
	part := strings.Split(s, ":")
	n := len(part)
	if n > 0 {
		r.Resource = part[0]
	}
	if n > 1 {
		r.Method = part[1]
	}
	return
}

// Match returns whether the scope is a match.
func (r *BaseScope) Match(resource string, method string) (b bool) {
	b = (r.Resource == "*" || strings.EqualFold(r.Resource, resource)) &&
		(r.Method == "*" || strings.EqualFold(r.Method, method))
	return
}

// String representations of the scope.
func (r *BaseScope) String() (s string) {
	s = strings.Join([]string{r.Resource, r.Method}, ":")
	return
}
