package tagcategory

import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	RichClient  *binding.RichClient
	TagCategory binding.TagCategory
)

func init() {
	// Prepare RichClient and login to Hub API (configured from env variables).
	RichClient = client.PrepareRichClient()

	// Shortcut for TagCategory-related RichClient methods.
	TagCategory = RichClient.TagCategory
}
