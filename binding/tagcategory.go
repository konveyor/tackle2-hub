package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// TagCategory API.
type TagCategory struct {
	// hub API client.
	Client *Client
}

//
// Create a TagCategory.
func (h *TagCategory) Create(r *api.TagCategory) (err error) {
	err = h.Client.Post(api.TagCategoriesRoot, &r)
	return
}

//
// Get a TagCategory by ID.
func (h *TagCategory) Get(id uint) (r *api.TagCategory, err error) {
	r = &api.TagCategory{}
	path := Path(api.TagCategoryRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List TagCategories.
func (h *TagCategory) List() (list []api.TagCategory, err error) {
	list = []api.TagCategory{}
	err = h.Client.Get(api.TagCategoriesRoot, &list)
	return
}

//
// Update a TagCategory.
func (h *TagCategory) Update(r *api.TagCategory) (err error) {
	path := Path(api.TagCategoryRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete a TagCategory.
func (h *TagCategory) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.TagCategoryRoot).Inject(Params{api.ID: id}))
	return
}
