package client

import (
	"fmt"

	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/settings"
)


func PrepareRichClient() (richClient *binding.RichClient) {
		// Prepare RichClient and login to Hub API
		richClient = binding.New(settings.Settings.Addon.Hub.URL)
		err := richClient.Login(settings.Settings.Auth.Keycloak.Admin.User, settings.Settings.Auth.Keycloak.Admin.Pass)
		if err != nil {
		  panic(fmt.Sprintf("Cannot login to API: %v.", err.Error()))
		}
	
		// Disable HTTP requests retry for network-related errors to fail quickly.
		richClient.Client().Retry = 0

		return
}