package tagcategory

import (
	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = c.Client
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
// Create a TagCategory.
func Create(r *api.TagCategory) (err error) {
	err = Client.Post(api.TagCategoriesRoot, &r)
	return
}

//
// Retrieve the TagCategory.
func Get(r *api.TagCategory) (err error) {
	err = Client.Get(c.Path(api.TagCategoriesRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Update the TagCategory.
func Update(r *api.TagCategory) (err error) {
	err = Client.Put(c.Path(api.TagCategoryRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Delete the TagCategory.
func Delete(r *api.TagCategory) (err error) {
	err = Client.Delete(c.Path(api.TagCategoryRoot, c.Params{api.ID: r.ID}))
	return
}

//
// List TagCategories.
func List(r []*api.TagCategory) (err error) {
	err = Client.Get(api.TagCategoriesRoot, &r)
	return
}