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
	EnvAddonAccessSecret    = "ADDON_ACCESS_SECRET"
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
	// Addon API access secret
	AddonAccessSecret string
}

func (r *Auth) Load() (err error) {
	var found bool
	r.Required = getEnvBool(EnvAuthRequired, true)
	r.Keycloak.Host, found = os.LookupEnv(EnvKeycloakHost)
	if r.Required && !found {
		panic("KEYCLOAK_HOST required")
	}
	r.Keycloak.Realm, found = os.LookupEnv(EnvKeycloakRealm)
	if r.Required && !found {
		panic("KEYCLOAK_REALM required")
	}
	r.Keycloak.ClientID, found = os.LookupEnv(EnvKeycloakClientID)
	if r.Required && !found {
		panic("KEYCLOAK_CLIENT_ID required")
	}
	r.Keycloak.ClientSecret, found = os.LookupEnv(EnvKeycloakClientSecret)
	if r.Required && !found {
		panic("KEYCLOAK_CLIENT_SECRET required")
	}
	r.AddonAccessSecret, found = os.LookupEnv(EnvAddonAccessSecret)
	if r.Required && !found {
		panic("ADDON_ACCESS_SECRET required")
	}
	return
}
