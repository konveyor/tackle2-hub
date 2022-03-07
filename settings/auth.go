package settings

import (
	"os"
)

//
// Environment variables
const (
	EnvAuthRequired         = "AUTH_REQUIRED"
	EnvKeycloakHost         = "KEYCLOAK_HOST"
	EnvKeycloakRealm        = "KEYCLOAK_REALM"
	EnvKeycloakClientID     = "KEYCLOAK_CLIENT_ID"
	EnvKeycloakClientSecret = "KEYCLOAK_CLIENT_SECRET"
	EnvAddonToken           = "ADDON_TOKEN"
)

//
// Defaults
const (
	KeycloakHost         = "https://localhost:8081"
	KeycloakRealm        = "tackle"
	KeycloakClientID     = "tackle-hub"
	KeycloakClientSecret = "tackle"
	AddonToken           = "tackle"
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
	var found bool
	r.Required = getEnvBool(EnvAuthRequired, false)
	if !r.Required {
		return
	}
	r.Keycloak.Host, found = os.LookupEnv(EnvKeycloakHost)
	if !found {
		r.Keycloak.Host = KeycloakHost
	}
	r.Keycloak.Realm, found = os.LookupEnv(EnvKeycloakRealm)
	if !found {
		r.Keycloak.Realm = KeycloakRealm
	}
	r.Keycloak.ClientID, found = os.LookupEnv(EnvKeycloakClientID)
	if !found {
		r.Keycloak.ClientID = KeycloakClientID
	}
	r.Keycloak.ClientSecret, found = os.LookupEnv(EnvKeycloakClientSecret)
	if !found {
		r.Keycloak.ClientSecret = KeycloakClientSecret
	}
	r.AddonToken, found = os.LookupEnv(EnvAddonToken)
	if !found {
		r.AddonToken = AddonToken
	}
	return
}
