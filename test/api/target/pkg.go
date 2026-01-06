package target

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Target     binding2.Target
	RuleSet    binding2.RuleSet
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Target = RichClient.Target
	RuleSet = RichClient.RuleSet
}
