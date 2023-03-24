package client

import (
	"fmt"

	"github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/settings"
)

var Client *addon.Client

func init() {
	var err error
	Client, err = New()
	if err != nil {
		panic(fmt.Sprintf("Error: Cannot setup API client: %v.", err.Error()))
	}
}

//
// Create new Hub client with login.
// Configured with environment variables HUB_BASE_URL, KEYCLOAK_ADMIN_USER, KEYCLOAK_ADMIN_PASS.
func New() (client *addon.Client, err error) {
	baseUrl := settings.Settings.Addon.Hub.URL
	login := api.Login{User: settings.Settings.Auth.Keycloak.Admin.User, Password: settings.Settings.Auth.Keycloak.Admin.Pass}

	// Setup client.
	client = addon.NewClient(baseUrl, "")

	// Login.
	err = client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return
	}
	client.SetToken(login.Token)
	return
}
