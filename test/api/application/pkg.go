package application

import (
	"fmt"

	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/settings"
)

var (
	Client *binding.Client
	RichClient *binding.RichClient
	Application binding.Application
)


func init() {
	// Prepare RichClient and login to Hub API
	RichClient = binding.New(settings.Settings.Addon.Hub.URL)
	err := RichClient.Login(settings.Settings.Auth.Keycloak.Admin.User, settings.Settings.Auth.Keycloak.Admin.Pass)
	if err != nil {
	  panic(fmt.Sprintf("Cannot login to API: %v.", err.Error()))
	}

	// Access REST client directly (some test API call need it)
	Client = RichClient.Client()

	// Disable HTTP requests retry for network-related errors to fail quickly.
	Client.Retry = 0

	// Shortcut for Application-related RichClient methods.
	Application = RichClient.Application
	
}
