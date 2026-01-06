package setting

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Setting    binding2.Setting
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Setting-related RichClient methods.
	Setting = RichClient.Setting
}
