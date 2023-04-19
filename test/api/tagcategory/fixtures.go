package tagcategory

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Set of valid TagCategories resources for tests and reuse.
func Samples() (samples map[string]api.TagCategory) {
	samples = map[string]api.TagCategory{
		"Test OS": {
			Name:  "Test OS",
			Color: "#dd0000",
		},
		"Test Language": {
			Name:  "Test Language",
			Color: "#0000dd",
		},
	}
	return
}

// Create a TagCategory.
func Create(r *api.TagCategory) (err error) {
	err = Client.Post(api.TagCategoriesRoot, &r)
	return
}

// Retrieve the TagCategory.
func Get(r *api.TagCategory) (err error) {
	err = Client.Get(client.Path(api.TagCategoriesRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the TagCategory.
func Update(r *api.TagCategory) (err error) {
	err = Client.Put(client.Path(api.TagCategoryRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the TagCategory.
func Delete(r *api.TagCategory) (err error) {
	err = Client.Delete(client.Path(api.TagCategoryRoot, client.Params{api.ID: r.ID}))
	return
}

// List TagCategories.
func List(r []*api.TagCategory) (err error) {
	err = Client.Get(api.TagCategoriesRoot, &r)
	return
}
