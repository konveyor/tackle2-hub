package stakeholder

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	Stakeholder binding.Stakeholder
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Stakeholder-related RichClient methods.
	Stakeholder = RichClient.Stakeholder
}
