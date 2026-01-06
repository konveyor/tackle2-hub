package businessservice

import (
	binding2 "github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/internal/test/api/client"
)

var (
	RichClient      *binding2.RichClient
	BusinessService binding2.BusinessService
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for BusinessService-related RichClient methods.
	BusinessService = RichClient.BusinessService
}
