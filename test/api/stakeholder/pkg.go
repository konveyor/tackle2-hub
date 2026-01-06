package stakeholder

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding2.RichClient
	Stakeholder binding2.Stakeholder
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Stakeholder-related RichClient methods.
	Stakeholder = RichClient.Stakeholder
}
