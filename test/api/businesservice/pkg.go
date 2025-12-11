package businessservice

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient      *binding.RichClient
	BusinessService binding.BusinessService
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for BusinessService-related RichClient methods.
	BusinessService = RichClient.BusinessService
}
