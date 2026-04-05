package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"gorm.io/gorm"
)

var (
	// Log logger.
	Log = logr.New("auth", Settings.Log.Auth)
	// Hub provider.
	Hub Provider
)

func init() {
	Hub = &NoAuth{}
}

// New returns an auth provider.
func New(db *gorm.DB) (p Provider, err error) {
	builtin, err := NewBuiltin(db)
	if err != nil {
		return
	}
	p = &NoAuth{handler: builtin.Handler()}
	if Settings.Auth.Required {
		p = builtin
	}
	return
}

// Provider provides RBAC.
type Provider interface {
	// Grant the key request.
	Grant(kr KeyRequest) (key APIKey, err error)
	// Authenticate the request.
	Authenticate(r *Request) (jwToken *jwt.Token, err error)
	// Scopes extracts a list of scopes from the token.
	Scopes(jwToken *jwt.Token) []Scope
	// User extracts the user from token.
	User(jwToken *jwt.Token) (user string)
	// Handler returns an OIDC handler.
	Handler() (h http.Handler)
	// Delete api key.
	Delete(digest string) (err error)
}

// APIKey authentication key.
type APIKey struct {
	User       string
	Digest     string
	Secret     string
	Scopes     []string
	Expiration time.Time
}

// JWT Claims - Standard claims.
const (
	ClaimSub   = "sub"   // Subject
	ClaimScope = "scope" // Scope
	ClaimExp   = "exp"   // Expiration Time
	ClaimIss   = "iss"   // Issuer
	ClaimAud   = "aud"   // Audience
	ClaimIat   = "iat"   // Issued At
	ClaimJti   = "jti"   // JWT ID
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

// notFound returns goidc.ErrNotFound when
// err IsA gorm.ErrRecordNotFound.
// Else, wrapped.
func notFound(err error) (e2 error) {
	if err == nil {
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		e2 = goidc.ErrNotFound
	} else {
		e2 = liberr.Wrap(err)
	}
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

// hashSecret hashes an API key secret using HMAC-SHA256.
func hashSecret(plainSecret string) (hashed string) {
	h := hmac.New(sha256.New, []byte(Settings.Auth.APIKey.Secret))
	h.Write([]byte(plainSecret))
	hash := h.Sum(nil)
	hashed = base64.StdEncoding.EncodeToString(hash)
	return
}
