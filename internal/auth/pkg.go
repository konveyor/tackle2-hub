package auth

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/auth/cache"
	"github.com/konveyor/tackle2-hub/internal/auth/seed"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
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
	Settings        = &settings.Settings
	Log             = logr.New("auth", Settings.Log.Auth)
	reloadMutex     sync.Mutex
	currentProvider atomic.Value
	currentDomain   atomic.Value
)

// Idp returns the current auth provider.
func Idp() (p Provider) {
	v := currentProvider.Load()
	if v != nil {
		p = v.(Provider)
	}
	return
}

// SetIdp sets the auth provider.
func SetIdp(p Provider) {
	if p != nil {
		Log.Info("Provider updated")
		currentProvider.Store(p)
	}
}

// Domain returns the current tenant.
func Domain() (d *Tenant) {
	v := currentDomain.Load()
	if v != nil {
		d = v.(*Tenant)
	}
	return
}

// SetDomain sets the tenant.
func SetDomain(d *Tenant) {
	if d != nil {
		Log.Info("Domain updated:\n" + d.String())
		currentDomain.Store(d)
	}
}

// Reload reloads the domain and auth provider.
// Preserves registered resources from the current domain.
func Reload(db *gorm.DB, client k8sClient.Client) (err error) {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	d := Domain()
	tenant := NewTenant(db, client)
	tenant.resources = d.resources
	err = tenant.Load()
	if err != nil {
		return
	}
	err = tenant.Seed()
	if err != nil {
		return
	}
	SetDomain(tenant)
	p, err := New(db)
	if err != nil {
		return
	}
	SetIdp(p)
	Log.Info("Reloaded.")
	return
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
