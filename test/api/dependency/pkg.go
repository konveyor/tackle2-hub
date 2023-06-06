package dependency

import (
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Dependency binding.Dependency
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Dependencyrelated RichClient methods.
	Dependency = RichClient.Dependency
}
