package jobfunction

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	JobFunction binding.JobFunction
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for JobFunction-related RichClient methods.
	JobFunction = RichClient.JobFunction
}
