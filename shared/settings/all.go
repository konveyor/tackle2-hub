package settings

import (
	"gopkg.in/yaml.v2"
)

var Settings TackleSettings

func init() {
	err := Settings.Load()
	if err != nil {
		panic(err)
	}
}

// TackleSettings project settings.
type TackleSettings struct {
	Hub
	Auth
	Metrics
	Addon
}

func (r *TackleSettings) Load() (err error) {
	err = r.Hub.Load()
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
	err = r.Addon.Load()
	if err != nil {
		return
	}
	return
}

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
