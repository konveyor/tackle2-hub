package review

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	Review      binding.Review
	Application binding.Application
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for Review-related RichClient methods.
	Review = RichClient.Review

	// Shortcut for Application-related RichClient methods.
	Application = RichClient.Application
}
