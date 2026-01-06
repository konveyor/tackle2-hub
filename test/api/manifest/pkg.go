package manifest

import (
	"github.com/konveyor/tackle2-hub/api"
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Manifest   binding2.Manifest
)
var (
	Decrypted = binding2.Param{Key: api.Decrypted, Value: "1"}
	Injected  = binding2.Param{Key: api.Injected, Value: "1"}
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for RuleSet-related RichClient methods.
	Manifest = RichClient.Manifest
}
