package tracker

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Tracker    binding.Tracker
	Identity   binding.Identity
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Tracker-related RichClient methods.
	Tracker = RichClient.Tracker

	// Shortcut for Identity-related RichClient methods.
	Identity = RichClient.Identity
}
