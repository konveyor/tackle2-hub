package tag

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Tag        binding2.Tag
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Tag-related RichClient methods.
	Tag = RichClient.Tag
}
