package ticket

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
)

var (
	RichClient  *binding2.RichClient
	Ticket      binding2.Ticket
	Tracker     binding2.Tracker
	Identity    binding2.Identity
	Application binding2.Application
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
