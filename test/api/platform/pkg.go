package manifest

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Platform   binding.Platform
)
var (
	Decrypted = binding.Param{Key: api.Decrypted, Value: "1"}
	Injected  = binding.Param{Key: api.Injected, Value: "1"}
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Platform = RichClient.Platform
}
