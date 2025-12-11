package task

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient *binding.RichClient
	Task       binding.Task
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Task-related RichClient methods.
	Task = RichClient.Task
}
