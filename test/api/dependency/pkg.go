package dependency

import (
	binding2 "github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding2.RichClient
	Dependency  binding2.Dependency
	Application binding2.Application
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Dependency related RichClient methods.
	Dependency = RichClient.Dependency

	//Shortcut for Application related RichClient methods.
	Application = RichClient.Application
}
