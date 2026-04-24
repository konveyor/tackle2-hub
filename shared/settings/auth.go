package settings

import (
	"os"
	"strings"
	"time"

	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables
const (
	EnvAuthRequired         = "AUTH_REQUIRED"
	EnvAPIKeySecret         = "APIKEY_SECRET"
	EnvAPIKeyLifespan       = "APIKEY_LIFESPAN"
	EnvAPIKeyCacheLifespan  = "APIKEY_CACHE_LIFESPAN"
	EnvIssuerURL            = "OIDC_ISSUER_URL"
	EnvTokenKey             = "ADDON_TOKEN" // Deprecated
	EnvTokenLifespan        = "OIDC_TOKEN_LIFESPAN"
	EnvRefreshTokenLifespan = "OIDC_REFRESH_TOKEN_LIFESPAN"
	EnvClientRedirectURIs   = "OIDC_CLIENT_REDIRECT_URIS"
	EnvClientID             = "OIDC_CLIENT_ID"
	EnvClientName           = "OIDC_CLIENT_NAME"
	EnvClientSecret         = "OIDC_CLIENT_SECRET"
	EnvKeyRotation          = "OIDC_KEY_ROTATION"
	EnvIdpEnabled           = "IDP_ENABLED"
	EnvIdpName              = "IDP_NAME"
	EnvIdpIssuerURL         = "IDP_ISSUER_URL"
	EnvIdpClientID          = "IDP_CLIENT_ID"
	EnvIdpClientSecret      = "IDP_CLIENT_SECRET"
	EnvIdpRedirectURI       = "IDP_REDIRECT_URI"
	EnvIdpScopes            = "IDP_SCOPES"
)

type Auth struct {
	// Auth required
	Required bool
	// APIKey settings.
	APIKey struct {
		Secret        string
		CacheLifespan time.Duration
		Lifespan      time.Duration
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
	// OIDC client settings.
	Client struct {
		ID           string
		Name         string
		Secret       string
		RedirectURIs []string
	}
	// IDP (identity-provider) settings
	Idp struct {
		Enabled      bool
		Name         string
		IssuerURL    string
		ClientID     string
		ClientSecret string
		RedirectURI  string
		Scopes       []string
	}
}

func (r *Auth) Load() (err error) {
	// API-Key
	r.Required = env.GetBool(EnvAuthRequired, false)
	r.APIKey.Secret = env.Get(EnvAPIKeySecret, "tackle")
	r.APIKey.CacheLifespan = env.GetMinute(EnvAPIKeyCacheLifespan, 5) // minutes: 5
	r.APIKey.Lifespan = env.GetHour(EnvAPIKeyLifespan, 10*24*365)     // hours: 10 years.
	// Token
	r.Token.Key = env.Get(EnvTokenKey, "tackle")
	r.Token.Lifespan = env.GetSecond(EnvTokenLifespan, 300)                   // seconds: 5 minutes.
	r.Token.RefreshLifespan = env.GetSecond(EnvRefreshTokenLifespan, 48*3600) // seconds: 2 days.
	// OIDC Provider
	r.IssuerURL, _ = os.LookupEnv(EnvIssuerURL)
	r.Key.Rotation = env.GetDay(EnvKeyRotation, 90)
	r.Client.ID = env.Get(EnvClientID, "web-ui")
	r.Client.Name = env.Get(EnvClientName, "Web UI")
	r.Client.Secret, _ = os.LookupEnv(EnvClientSecret)
	s, found := os.LookupEnv(EnvClientRedirectURIs)
	if found {
		r.Client.RedirectURIs = strings.Split(s, ",")
		for i := range r.Client.RedirectURIs {
			r.Client.RedirectURIs[i] = strings.TrimSpace(r.Client.RedirectURIs[i])
		}
	}
	// Remote IDP Endpoint.
	r.Idp.Enabled = env.GetBool(EnvIdpEnabled, false)
	r.Idp.Name = env.Get(EnvIdpName, "tackle")
	r.Idp.IssuerURL, _ = os.LookupEnv(EnvIdpIssuerURL)
	r.Idp.ClientID, _ = os.LookupEnv(EnvIdpClientID)
	r.Idp.ClientSecret, _ = os.LookupEnv(EnvIdpClientSecret)
	r.Idp.RedirectURI, _ = os.LookupEnv(EnvIdpRedirectURI)
	s, found = os.LookupEnv(EnvIdpScopes)
	if found {
		r.Idp.Scopes = strings.Split(s, ",")
		for i := range r.Idp.Scopes {
			r.Idp.Scopes[i] = strings.TrimSpace(r.Idp.Scopes[i])
		}
	} else {
		r.Idp.Scopes = []string{
			"offline_access",
			"openid",
			"profile",
			"email",
		}
	}
	return
}
