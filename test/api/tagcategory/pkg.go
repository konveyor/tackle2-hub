package tagcategory

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

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
