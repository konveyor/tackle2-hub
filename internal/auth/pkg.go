package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jortel/go-utils/logr"
	"gorm.io/gorm"
)

const (
	KindAccessToken = "access"
	KindAuthCode    = "authCode"
	KindAPIKey      = "api-key"
	KindDevice      = "devCode"
)

var (
	// Log logger.
	Log = logr.New("auth", Settings.Log.Auth)
	// IdP provider.
	IdP Provider
)

func init() {
	IdP = &NoAuth{}
}

// New returns an auth provider.
func New(db *gorm.DB) (p Provider, err error) {
	builtin, err := NewBuiltin(db)
	if err != nil {
		return
	}
	p = NewNoAuth(builtin)
	if Settings.Auth.Required {
		p = builtin
	}
	return
}

// Provider provides RBAC.
type Provider interface {
	// Cache returns the provider cache.
	Cache() *Cache
	// Login begin OIDC auth.
	Login(w http.ResponseWriter, r *http.Request, reqId string) (err error)
	// NewPAT creates a new personal access token.
	NewPAT(subject string, lifespan time.Duration) (token Token, err error)
	// NewTaskToken creates a new api-key.
	NewTaskToken(taskId uint) (token Token, err error)
	// Revoke a token.
	Revoke(tokenId uint) (err error)
	// Authenticate the request.
	Authenticate(r *Request) (jwToken *jwt.Token, err error)
	// Scopes extracts a list of scopes from the token.
	Scopes(jwToken *jwt.Token) []Scope
	// User extracts the user from token.
	User(jwToken *jwt.Token) (user string)
	// Handler returns an OIDC handler.
	Handler() (h http.Handler)
	// IdpHandler returns the external IdP handler.
	IdpHandler() (h *IdpHandler)
}

// JWT Claims - Standard claims.
const (
	ClaimSub   = "sub"   // Subject
	ClaimScope = "scope" // Scope
	ClaimExp   = "exp"   // Expiration Time
	ClaimIss   = "iss"   // Issuer
	ClaimAud   = "aud"   // Audience
)

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

// hasExpiredIdentity returns true when the token references and expired IpP identity.
func (t *Token) hasExpiredIdentity() (expired bool) {
	id := t.IdpIdentity
	if id == nil {
		return
	}
	expired = id.Expiration.Before(time.Now())
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

// asTime returns a time.Time for unix time.
func asTime(n int) (t time.Time) {
	t = time.Unix(int64(n), 0)
	t = t.UTC()
	return
}

// asInt returns unix time for time.Time.
func asInt(t time.Time) (i int) {
	t = t.UTC()
	i = int(t.Unix())
	return
}
