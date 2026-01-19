package dependency

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/application"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	Dependency  binding.Dependency
	Application application.Application
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Dependency related RichClient methods.
	Dependency = RichClient.Dependency

	//Shortcut for Application related RichClient methods.
	Application = RichClient.Application
}
