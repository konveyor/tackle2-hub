package tagcategory

import (
	"testing"

	"github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = client.Client
)

//
// Set of valid TagCategories resources for tests and reuse.
func Samples() (samples []api.TagCategory) {
	samples = []api.TagCategory{
		{
			Name:  "Test OS",
			Color: "#dd0000",
		},
		{
			Name:  "Test Language",
			Color: "#0000dd",
		},
	}
	return
}

//
// Create.
func Create(t *testing.T, r *api.TagCategory) {
	err := Client.Post(api.TagCategoriesRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error())
	}
}

//
// Delete.
func Delete(t *testing.T, r *api.TagCategory) {
	err := Client.Delete(addon.Params{api.ID: r.ID}.Inject(api.TagCategoryRoot))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
