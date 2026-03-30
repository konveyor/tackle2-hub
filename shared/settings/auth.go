package settings

import (
	"os"
	"strings"

	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables
const (
	EnvAuthRequired    = "AUTH_REQUIRED"
	EnvBuiltinTokenKey = "ADDON_TOKEN"
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
	// Token settings for builtin provider.
	Token struct {
		Key string
	}
	// IDP settings
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
	if !r.Required {
		return
	}
	r.Token.Key = env.Get(EnvBuiltinTokenKey, "konveyor")
	r.Idp.Enabled = env.GetBool(EnvIdpEnabled, false)
	r.Idp.Name = env.Get(EnvIdpName, "tackle")
	r.Idp.IssuerURL, _ = os.LookupEnv(EnvIdpIssuerURL)
	r.Idp.ClientID, _ = os.LookupEnv(EnvIdpClientID)
	r.Idp.ClientSecret, _ = os.LookupEnv(EnvIdpClientSecret)
	r.Idp.RedirectURI = env.Get(EnvIdpRedirectURI, "http://localhost:8080/idp/callback")
	s, found := os.LookupEnv(EnvIdpScopes)
	if found && s != "" {
		r.Idp.Scopes = strings.Split(s, ",")
	} else {
		r.Idp.Scopes = []string{
			"openid",
			"profile",
			"email"}
	}
	return
}
