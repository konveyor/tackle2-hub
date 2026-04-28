package seed

import (
	"os"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gopkg.in/yaml.v3"
)

var (
	Settings = &settings.Settings
)

// IdP configuration loaded from YAML.
type IdP struct {
	Clients    []Client   `yaml:"clients"`
	Federation Federation `yaml:"federation"`
}

// Load reads and parses IdP configuration from a YAML file.
// Injects required clients (device-verifier) based on runtime settings.
func (r *IdP) Load(path string) (err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = yaml.Unmarshal(content, r)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.Clients = append(
		r.Clients, Client{
			ClientId:        "device-verifier",
			ClientSecret:    "",
			ApplicationType: "web",
			Grants:          []string{"authorization_code"},
			RedirectURIs:    []string{Settings.Auth.IssuerWithPath(api.AuthDevAuthCallback)},
			Scopes:          []string{"openid"},
		})

	return
}

// Client configuration.
type Client struct {
	ClientId        string   `yaml:"clientId"`
	ClientSecret    string   `yaml:"clientSecret"`
	ApplicationType string   `yaml:"applicationType"`
	Grants          []string `yaml:"grants"`
	RedirectURIs    []string `yaml:"redirectURIs"`
	Scopes          []string `yaml:"scopes"`
}

// Federation configuration.
type Federation struct {
	Ldap interface{} `yaml:"ldap"`
	Idp  struct {
		Name         string   `yaml:"name"`
		Issuer       string   `yaml:"issuer"`
		ClientId     string   `yaml:"clientId"`
		ClientSecret string   `yaml:"clientSecret"`
		Scopes       []string `yaml:"scopes"`
	}
}
