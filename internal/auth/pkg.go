package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/auth/cache"
	"github.com/konveyor/tackle2-hub/internal/auth/seed"
	as "github.com/konveyor/tackle2-hub/internal/auth/settings"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
)

const (
	ADMIN = seed.ADMIN
)

const (
	KindAccessToken = cache.KindAccessToken
	KindAuthCode    = cache.KindAuthCode
	KindAPIKey      = cache.KindAPIKey
)

const (
	IdentityKindOpenid = cache.IdentityKindOpenid
	IdentityKindLDAP   = cache.IdentityKindLDAP
)

const (
	DevVerifierClientId = "device-verifier"
)

const (
	AuthRequestId = "authRequestId"
)

var (
	Settings  = &settings.Settings
	Log       = logr.New("auth", Settings.Log.Auth)
	federated = &as.Federated{}
	IdP       Provider
)

// New returns an auth provider.
func New(db *gorm.DB) (p Provider, err error) {
	err = federated.Load(Settings.Namespace)
	if err != nil {
		return
	}
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
	// Ready notification of an incoming request.
	Ready(r *http.Request)
	// Cache returns the provider cache.
	Cache() *Cache
	// Login begin OIDC auth.
	Login(w http.ResponseWriter, r *http.Request, reqId string) (err error)
	// NewToken creates a new personal access token.
	NewToken(subject string, lifespan time.Duration) (token Token, err error)
	// TaskGrant creates a new api-key.
	TaskGrant(task *Task) (token Token, err error)
	// TaskRevoke revokes task tokens.
	TaskRevoke(taskId uint)
	// Revoke a token.
	Revoke(tokenId uint) (err error)
	// Authenticate the request.
	Authenticate(r *Request) (jwToken *jwt.Token, err error)
	// Scopes extracts a list of scopes from the token.
	Scopes(jwToken *jwt.Token) []Scope
	// User extracts the user from token.
	User(jwToken *jwt.Token) (user string)
	// Subject extracts the subject from the token.
	Subject(jwToken *jwt.Token) (subject string)
	// Handler returns an OIDC handler.
	Handler() (h http.Handler)
	// IdpHandler returns the external IdP handler.
	IdpHandler() (h *FedIdpHandler)
	// DagHandler returns the device access grant handler.
	DagHandler() (h *DagHandler)
}

// JWT Claims - Standard claims.
const (
	ClaimSub   = "sub"   // Subject
	ClaimScope = "scope" // Scope
	ClaimExp   = "exp"   // Expiration Time
	ClaimIss   = "iss"   // Issuer
	ClaimAud   = "aud"   // Audience
	ClaimIat   = "iat"   // Issued At
	ClaimId    = "jti"   // Token id.
)

// cache aliases

type RsaKey = model.RsaKey
type Cache = cache.Cache
type Tx = cache.Tx
type Model = cache.Model
type User = cache.User
type Role = cache.Role
type Token = cache.Token
type Identity = cache.Identity
type Subject = cache.Subject
type Task = cache.Task
type Permission = cache.Permission
type Grant = cache.Grant

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

func uniqueStrings(items []string) (result []string) {
	seen := make(map[string]bool, len(items))
	result = make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return
}

type IdpClient = cache.IdpClient
