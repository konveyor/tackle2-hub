package settings

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

var Settings TackleSettings

type TackleSettings struct {
	Hub
	Metrics
	Addon
	Auth
}

func (r *TackleSettings) Load() (err error) {
	err = r.Hub.Load()
	if err != nil {
		return
	}
	err = r.Addon.Load()
	if err != nil {
		return
	}
	err = r.Auth.Load()
	if err != nil {
		return
	}
	err = r.Metrics.Load()
	if err != nil {
		return
	}
	return
}

// String returns a YAML representation.
// Redacted as needed.
func (r TackleSettings) String() (s string) {
	redacted := "********"
	r.Encryption.Passphrase = redacted
	r.Auth.Keycloak.ClientSecret = redacted
	r.Auth.Keycloak.Admin.Pass = redacted
	r.Auth.Keycloak.Admin.User = redacted
	b, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}
	s = string(b)
	return
}

// Get boolean.
func getEnvBool(name string, def bool) bool {
	boolean := def
	if s, found := os.LookupEnv(name); found {
		parsed, err := strconv.ParseBool(s)
		if err == nil {
			boolean = parsed
		}
	}

	return boolean
}
