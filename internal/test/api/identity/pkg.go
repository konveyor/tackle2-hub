package identity

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Identity   binding2.Identity
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Identity-related RichClient methods.
	Identity = RichClient.Identity
}
