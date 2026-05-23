package settings

import (
	"net/url"
	"time"

	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables
const (
	EnvAuthEnabled          = "AUTH_ENABLED"
	EnvAuthRequired         = "AUTH_REQUIRED"
	EnvAPIKeySecret         = "APIKEY_SECRET"
	EnvAPIKeyLifespan       = "APIKEY_LIFESPAN"
	EnvCacheLifespan        = "AUTH_CACHE_LIFESPAN"
	EnvBasicAuthLifespan    = "BASIC_AUTH_LIFESPAN"
	EnvTokenKey             = "ADDON_TOKEN" // Deprecated
	EnvOidcIssuer           = "OIDC_ISSUER"
	EnvTokenLifespan        = "OIDC_TOKEN_LIFESPAN"
	EnvRefreshTokenLifespan = "OIDC_REFRESH_TOKEN_LIFESPAN"
	EnvKeyRotation          = "OIDC_KEY_ROTATION"
	EnvRedirectURIWebUI     = "OIDC_REDIRECT_URI_WEBUI"
)

type Auth struct {
	// Auth enabled
	Enabled bool
	// Auth required
	Required bool
	// Cache
	CacheLifespan time.Duration
	// BasicAuth cache lifespan
	BasicAuthLifespan time.Duration
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
	// RedirectURI settings for OIDC clients.
	RedirectURI struct {
		WebUI string
	}
}

func (r *Auth) Load() (err error) {
	r.Enabled = env.GetBool(EnvAuthEnabled, true)
	r.Required = env.GetBool(EnvAuthRequired, false)
	r.CacheLifespan = env.GetMinute(EnvCacheLifespan, 5)
	r.BasicAuthLifespan = env.GetSecond(EnvBasicAuthLifespan, 60) // second: 1 minute.
	r.APIKey.Secret = env.Get(EnvAPIKeySecret, "tackle")
	r.APIKey.Lifespan = env.GetHour(EnvAPIKeyLifespan, 10*24*365) // hour: 10 years.
	r.Token.Key = env.Get(EnvTokenKey, "tackle")
	r.Token.Lifespan = env.GetSecond(EnvTokenLifespan, 300)                   // second: 5 minutes.
	r.Token.RefreshLifespan = env.GetSecond(EnvRefreshTokenLifespan, 48*3600) // second: 2 days.
	r.IssuerURL = env.Get(EnvOidcIssuer, "http://localhost:8080")
	r.Key.Rotation = env.GetDay(EnvKeyRotation, 90)
	r.RedirectURI.WebUI = env.Get(EnvRedirectURIWebUI, "")

	issuerURL, err := url.Parse(r.IssuerURL)
	if err == nil {
		if r.RedirectURI.WebUI == "" {
			issuerURL.Path = ""
			r.RedirectURI.WebUI = issuerURL.String()
		}
	} else {
		panic(err)
	}

	return
}

// AppendIssuer appends a path to the issuer URL.
func (r *Auth) AppendIssuer(path string) (s string) {
	issuerURL, _ := url.Parse(r.IssuerURL)
	joined, _ := url.JoinPath(issuerURL.Path, path)
	issuerURL.Path = joined
	s = issuerURL.String()
	return
}
