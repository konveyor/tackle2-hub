package client

import (
	"fmt"
	"os"

	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"k8s.io/utils/env"
)

const (
	Username = "HUB_USERNAME"
	Password = "HUB_PASSWORD"
)

// Create RichClient to interact with Hub API
// Parameters are read environment variables:
//
//	HUB_BASE_URL (required)
//	HUB_USERNAME, HUB_PASSWORD (optional, depends on Require Auth option in Konveyor installation)
func PrepareRichClient() (richClient *binding.RichClient) {
	// Prepare RichClient and login to Hub API
	richClient = binding.New(
		env.GetString(
			settings.EnvHubBaseURL,
			"http://localhost:8080"))
	err := richClient.Login(
		os.Getenv(Username),
		os.Getenv(Password))
	if err != nil {
		panic(fmt.Sprintf("Cannot login to API: %v.", err.Error()))
	}

	// Disable HTTP requests retry for network-related errors to fail quickly.
	richClient.Client.Retry = 1

	return
}
