package settings

import (
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
	EnvLdapAuthLifespan     = "LDAP_AUTH_LIFESPAN"
	EnvTokenKey             = "ADDON_TOKEN" // Deprecated
	EnvTokenLifespan        = "OIDC_TOKEN_LIFESPAN"
	EnvRefreshTokenLifespan = "OIDC_REFRESH_TOKEN_LIFESPAN"
	EnvKeyRotation          = "OIDC_KEY_ROTATION"
)

type Auth struct {
	// Auth enabled
	Enabled bool
	// Auth required
	Required bool
	// Cache lifespan.
	CacheLifespan time.Duration
	// LDAP auth lifespan.
	LdapAuthLifespan time.Duration
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
}

func (r *Auth) Load() (err error) {
	r.Enabled = env.GetBool(EnvAuthEnabled, true)
	r.Required = env.GetBool(EnvAuthRequired, false)
	r.CacheLifespan = env.GetMinute(EnvCacheLifespan, 5)       // minute: 5 minutes.
	r.LdapAuthLifespan = env.GetMinute(EnvLdapAuthLifespan, 5) // minute: 5 minutes.
	r.APIKey.Secret = env.Get(EnvAPIKeySecret, "tackle")
	r.APIKey.Lifespan = env.GetHour(EnvAPIKeyLifespan, 10*24*365) // hour: 10 years.
	r.Token.Key = env.Get(EnvTokenKey, "tackle")
	r.Token.Lifespan = env.GetSecond(EnvTokenLifespan, 300)                   // second: 5 minutes.
	r.Token.RefreshLifespan = env.GetSecond(EnvRefreshTokenLifespan, 48*3600) // second: 2 days.
	r.Key.Rotation = env.GetDay(EnvKeyRotation, 90)
	return
}
