package settings

import (
	"net/url"
	"os"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/env"
	"gopkg.in/yaml.v3"
)

// Environment variables
const (
	EnvAuthEnabled          = "AUTH_ENABLED"
	EnvAuthRequired         = "AUTH_REQUIRED"
	EnvAPIKeySecret         = "APIKEY_SECRET"
	EnvAPIKeyLifespan       = "APIKEY_LIFESPAN"
	EnvCacheLifespan        = "AUTH_CACHE_LIFESPAN"
	EnvTokenKey             = "ADDON_TOKEN" // Deprecated
	EnvTokenLifespan        = "OIDC_TOKEN_LIFESPAN"
	EnvRefreshTokenLifespan = "OIDC_REFRESH_TOKEN_LIFESPAN"
	EnvKeyRotation          = "OIDC_KEY_ROTATION"
	EnvAuthFile             = "AUTH_FILE"
)

type Auth struct {
	// Auth enabled
	Enabled bool
	// Auth required
	Required bool
	// Cache
	CacheLifespan time.Duration
	// APIKey settings.
	APIKey struct {
		Secret   string
		Lifespan time.Duration
	}
	// Token settings for builtin provider.
	Token struct {
		Key             string // Deprecated.
		Lifespan        time.Duration
		RefreshLifespan time.Duration
	}
	// RSaKey settings.
	Key struct {
		Rotation time.Duration
	}
	// OIDC Issuer
	IssuerURL string
	// OIDC Clients
	Clients []IdpClient
	// Federation settings
	Federation IdpFederation
}

func (r *Auth) Load() (err error) {
	r.Enabled = env.GetBool(EnvAuthEnabled, true)
	r.Required = env.GetBool(EnvAuthRequired, false)
	r.CacheLifespan = env.GetMinute(EnvCacheLifespan, 5)
	// API-Key
	r.APIKey.Secret = env.Get(EnvAPIKeySecret, "tackle")
	r.APIKey.Lifespan = env.GetHour(EnvAPIKeyLifespan, 10*24*365) // hour: 10 years.
	// Token
	r.Token.Key = env.Get(EnvTokenKey, "tackle")
	r.Token.Lifespan = env.GetSecond(EnvTokenLifespan, 300)                   // second: 5 minutes.
	r.Token.RefreshLifespan = env.GetSecond(EnvRefreshTokenLifespan, 48*3600) // second: 2 days.
	// OIDC Provider
	r.Key.Rotation = env.GetDay(EnvKeyRotation, 90)
	// IdP
	f := AuthFile{}
	f.Load()
	r.IssuerURL = f.IssuerURL
	r.Clients = f.Clients
	r.Federation = f.Federation

	if _, err := url.Parse(r.IssuerURL); err != nil {
		panic(err)
	}
	return
}

// IssuerWithPath returns the issuer URL with an alternate path.
func (r *Auth) IssuerWithPath(path string) (s string) {
	p, err := url.Parse(r.IssuerURL)
	if err != nil {
		return
	}
	p.Path = path
	s = p.String()
	return
}

// IdpClient settings.
type IdpClient struct {
	Id              string   `yaml:"id"`
	Secret          string   `yaml:"secret"`
	ApplicationType string   `yaml:"applicationType"`
	Grants          []string `yaml:"grants"`
	RedirectURIs    []string `yaml:"redirectURIs"`
	Scopes          []string `yaml:"scopes"`
}

// IdpFederation settings.
type IdpFederation struct {
	Enabled bool        `yaml:"enabled"`
	Ldap    interface{} `yaml:"ldap"`
	Idp     struct {
		Name         string   `yaml:"name"`
		Issuer       string   `yaml:"issuer"`
		ClientId     string   `yaml:"clientId"`
		ClientSecret string   `yaml:"clientSecret"`
		RedirectURI  string   `yaml:"redirectURI"`
		Scopes       []string `yaml:"scopes"`
	}
}

type AuthFile struct {
	IssuerURL  string        `yaml:"issuer"`
	Clients    []IdpClient   `yaml:"clients"`
	Federation IdpFederation `yaml:"federation"`
}

func (f *AuthFile) Load() {
	path := env.Get(EnvAuthFile, "/etc/hub/auth.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		err = liberr.Wrap(err)
		panic(err)
	}
	err = yaml.Unmarshal(content, f)
	if err != nil {
		err = liberr.Wrap(err)
		panic(err)
	}
}
