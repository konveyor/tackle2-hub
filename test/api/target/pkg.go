package target

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Target     binding.Target
	RuleSet    binding.RuleSet
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Target = RichClient.Target
	RuleSet = RichClient.RuleSet
}
