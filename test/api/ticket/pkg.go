package ticket

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	Ticket      binding.Ticket
	Tracker     binding.Tracker
	Identity    binding.Identity
	Application binding.Application
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Ticket-related RichClient methods.
	Ticket = RichClient.Ticket

	// Shortcut for Tracker-related RichClient methods.
	Tracker = RichClient.Tracker

	// Shortcut for Identity-related RichClient methods.
	Identity = RichClient.Identity

	// Shortcut for Application-related RichClient methods.
	Application = RichClient.Application
}
