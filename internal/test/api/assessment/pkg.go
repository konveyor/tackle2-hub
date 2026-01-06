package assessment

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
)

var (
	RichClient *binding2.RichClient
	Assessment binding2.Assessment
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
