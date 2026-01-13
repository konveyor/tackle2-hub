package application

import (
	binding2 "github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	Client      *binding2.Client
	RichClient  *binding2.RichClient
	Application binding2.Application
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Access REST client directly (some test API call need it)
	Client = RichClient.Client

	// Shortcut for Application-related RichClient methods.
	Application = RichClient.Application
}
