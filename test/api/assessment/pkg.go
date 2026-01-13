package assessment

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Assessment binding.Assessment
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Assessment-related RichClient methods.
	Assessment = RichClient.Assessment
}

func uint2ptr(u uint) *uint {
	return &u
}
