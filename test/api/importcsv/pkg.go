package importcsv

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/application"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	Client      *binding.Client
	Application application.Application
	Dependency  binding.Dependency
	Stakeholder binding.Stakeholder
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Access REST client directly
	Client = RichClient.Client

	// Access Application directly
	Application = RichClient.Application

	// Access Dependency directly
	Dependency = RichClient.Dependency

	// Access Stakeholder directly
	Stakeholder = RichClient.Stakeholder
}
