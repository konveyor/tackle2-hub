package tagcategory

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid TagCategories resources for tests and reuse.
var (
	TestOS = api.TagCategory{
		Name:  "Test OS",
		Color: "#dd0000",
	}
	TestLanguage = api.TagCategory{
		Name:  "Test Language",
		Color: "#0000dd",
	}
	Samples = []api.TagCategory{TestOS, TestLanguage}
)
