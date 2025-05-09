package manifest

import (
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Manifest   binding.Manifest
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Manifest = RichClient.Manifest
}
