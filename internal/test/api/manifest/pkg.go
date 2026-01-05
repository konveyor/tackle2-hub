package manifest

import (
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
)

var (
	RichClient *binding.RichClient
	Manifest   binding.Manifest
)
var (
	Decrypted = binding.Param{Key: api.Decrypted, Value: "1"}
	Injected  = binding.Param{Key: api.Injected, Value: "1"}
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Manifest = RichClient.Manifest
}
