package manifest

import (
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
	"github.com/konveyor/tackle2-hub/shared/binding"
)

var (
	RichClient *binding.RichClient
	Generator  binding.Generator
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Generator = RichClient.Generator
}
