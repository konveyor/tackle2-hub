package setting

import (
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
	"github.com/konveyor/tackle2-hub/shared/binding"
)

var (
	RichClient *binding.RichClient
	Setting    binding.Setting
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Setting-related RichClient methods.
	Setting = RichClient.Setting
}
