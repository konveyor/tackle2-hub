package settings

import (
	"os"
)

const (
	EnvAuthRequired         = "AUTH_REQUIRED"
	EnvKeycloakHost         = "KEYCLOAK_HOST"
	EnvKeycloakRealm        = "KEYCLOAK_REALM"
	EnvKeycloakClientID     = "KEYCLOAK_CLIENT_ID"
	EnvKeycloakClientSecret = "KEYCLOAK_CLIENT_SECRET"
	EnvAddonToken           = "ADDON_TOKEN"
)

type Auth struct {
	// Auth required
	Required bool
	// Keycloak client config
	Keycloak struct {
		Host         string
		Realm        string
		ClientID     string
		ClientSecret string
	}
	// Addon API access token
	AddonToken string
}

func (r *Auth) Load() (err error) {
	r.Required = getEnvBool(EnvAuthRequired, true)
	r.Keycloak.Host = os.Getenv(EnvKeycloakHost)
	r.Keycloak.Realm = os.Getenv(EnvKeycloakRealm)
	r.Keycloak.ClientID = os.Getenv(EnvKeycloakClientID)
	r.Keycloak.ClientSecret = os.Getenv(EnvKeycloakClientSecret)
	r.AddonToken = os.Getenv(EnvAddonToken)
	return
}
