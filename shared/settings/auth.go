package settings

import (
	"os"
	"strings"
	"time"

	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables
const (
	EnvAuthRequired    = "AUTH_REQUIRED"
	EnvAPIKeySecret    = "API_KEY_SECRET"
	EnvIssuerURL       = "OIDC_ISSUER_URL"
	EnvClientID        = "OIDC_CLIENT_ID"
	EnvClientName      = "OIDC_CLIENT_NAME"
	EnvClientSecret    = "OIDC_CLIENT_SECRET"
	EnvKeyRotation     = "OIDC_KEY_ROTATION"
	EnvIdpEnabled      = "IDP_ENABLED"
	EnvIdpName         = "IDP_NAME"
	EnvIdpIssuerURL    = "IDP_ISSUER_URL"
	EnvIdpClientID     = "IDP_CLIENT_ID"
	EnvIdpClientSecret = "IDP_CLIENT_SECRET"
	EnvIdpRedirectURI  = "IDP_REDIRECT_URI"
	EnvIdpScopes       = "IDP_SCOPES"
)

type Auth struct {
	// Auth required
	Required bool
	// APIKey secret for HMAC hashing
	APIKeySecret string
	// Token settings for builtin provider.
	Token struct {
		Key string
	}
	// Issuer
	IssuerURL string
	// Key settings.
	Key struct {
		Rotation time.Duration
	}
	// OIDC client settings.
	Client struct {
		ID     string
		Name   string
		Secret string
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
	r.Required = env.GetBool(EnvAuthRequired, false)
	r.APIKeySecret = env.Get(EnvAPIKeySecret, "tackle")
	r.IssuerURL, _ = os.LookupEnv(EnvIssuerURL)
	r.Key.Rotation = env.GetDay(EnvKeyRotation, 90)
	r.Client.ID = env.Get(EnvClientID, "main")
	r.Client.Name = env.Get(EnvClientName, "main")
	r.Client.Secret = env.Get(EnvClientSecret, "tackle")
	r.Idp.Enabled = env.GetBool(EnvIdpEnabled, false)
	r.Idp.Name = env.Get(EnvIdpName, "tackle")
	r.Idp.IssuerURL, _ = os.LookupEnv(EnvIdpIssuerURL)
	r.Idp.ClientID, _ = os.LookupEnv(EnvIdpClientID)
	r.Idp.ClientSecret, _ = os.LookupEnv(EnvIdpClientSecret)
	r.Idp.RedirectURI, _ = os.LookupEnv(EnvIdpRedirectURI)
	s, found := os.LookupEnv(EnvIdpScopes)
	if found && s != "" {
		r.Idp.Scopes = strings.Split(s, ",")
	} else {
		r.Idp.Scopes = []string{
			"openid",
			"profile",
			"email",
		}
	}
	return
}
