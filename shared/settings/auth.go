package settings

import (
	"os"

	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables
const (
	EnvAuthRequired          = "AUTH_REQUIRED"
	EnvKeycloakHost          = "KEYCLOAK_HOST"
	EnvKeycloakRealm         = "KEYCLOAK_REALM"
	EnvKeycloakClientID      = "KEYCLOAK_CLIENT_ID"
	EnvKeycloakClientSecret  = "KEYCLOAK_CLIENT_SECRET"
	EnvKeycloakAdminUser     = "KEYCLOAK_ADMIN_USER"
	EnvKeycloakAdminPass     = "KEYCLOAK_ADMIN_PASS"
	EnvKeycloakAdminRealm    = "KEYCLOAK_ADMIN_REALM"
	EnvKeycloakReqPassUpdate = "KEYCLOAK_REQ_PASS_UPDATE"
	EnvKeycloakAudience      = "KEYCLOAK_AUDIENCE"
	EnvBuiltinTokenKey       = "ADDON_TOKEN"
	EnvRolePath              = "ROLE_PATH"
	EnvUserPath              = "USER_PATH"
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
		Audience     string
		Admin        struct {
			User  string
			Pass  string
			Realm string
		}
		RequirePasswordUpdate bool
	}
	// Path to role yaml
	RolePath string
	// Path to user yaml
	UserPath string
	// Token settings for builtin provider.
	Token struct {
		Key string
	}
}

func (r *Auth) Load() (err error) {
	var found bool
	r.Required = env.GetBool(EnvAuthRequired, false)
	if !r.Required {
		return
	}
	r.Keycloak.Host, found = os.LookupEnv(EnvKeycloakHost)
	if !found {
		r.Keycloak.Host = "https://localhost:8081"
	}
	r.Keycloak.Realm, found = os.LookupEnv(EnvKeycloakRealm)
	if !found {
		r.Keycloak.Realm = "konveyor"
	}
	r.Keycloak.ClientID, found = os.LookupEnv(EnvKeycloakClientID)
	if !found {
		r.Keycloak.ClientID = "konveyor"
	}
	r.Keycloak.ClientSecret, found = os.LookupEnv(EnvKeycloakClientSecret)
	if !found {
		r.Keycloak.ClientSecret = ""
	}
	r.Keycloak.Audience, found = os.LookupEnv(EnvKeycloakAudience)
	if !found {
		r.Keycloak.Audience = r.Keycloak.ClientID
	}
	r.Keycloak.Admin.User, found = os.LookupEnv(EnvKeycloakAdminUser)
	if !found {
		r.Keycloak.Admin.User = "admin"
	}
	r.Keycloak.Admin.Pass, found = os.LookupEnv(EnvKeycloakAdminPass)
	if !found {
		r.Keycloak.Admin.Pass = "admin"
	}
	r.Keycloak.Admin.Realm, found = os.LookupEnv(EnvKeycloakAdminRealm)
	if !found {
		r.Keycloak.Admin.Realm = "master"
	}
	r.Keycloak.RequirePasswordUpdate = env.GetBool(EnvKeycloakReqPassUpdate, true)
	r.Token.Key, found = os.LookupEnv(EnvBuiltinTokenKey)
	if !found {
		r.Token.Key = "konveyor"
	}
	r.RolePath, found = os.LookupEnv(EnvRolePath)
	if !found {
		r.RolePath = "/tmp/roles.yaml"
	}
	r.UserPath, found = os.LookupEnv(EnvUserPath)
	if !found {
		r.UserPath = "/tmp/users.yaml"

	}
	return
}
