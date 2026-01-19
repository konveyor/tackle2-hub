package binding

import (
	"errors"

	"github.com/konveyor/tackle2-hub/shared/api"
)

// TagCategory API.
type TagCategory struct {
	client *Client
}

// Create a TagCategory.
func (h TagCategory) Create(r *api.TagCategory) (err error) {
	err = h.client.Post(api.TagCategoriesRoute, r)
	return
}

// Get a TagCategory by ID.
func (h TagCategory) Get(id uint) (r *api.TagCategory, err error) {
	r = &api.TagCategory{}
	path := Path(api.TagCategoryRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List TagCategories.
func (h TagCategory) List() (list []api.TagCategory, err error) {
	list = []api.TagCategory{}
	err = h.client.Get(api.TagCategoriesRoute, &list)
	return
}

// Update a TagCategory.
func (h TagCategory) Update(r *api.TagCategory) (err error) {
	path := Path(api.TagCategoryRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a TagCategory.
func (h TagCategory) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TagCategoryRoute).Inject(Params{api.ID: id}))
	return
}

// Find by name.
func (h TagCategory) Find(name string) (r *api.TagCategory, found bool, err error) {
	list := []api.TagCategory{}
	err = h.client.Get(
		api.TagCategoriesRoute,
		&list,
		Param{
			Key:   api.Name,
			Value: name,
		})
	if err != nil {
		return
	}
	if len(list) > 0 {
		found = true
		r = &list[0]
	}
	return
}

// Ensure a tag-type exists.
func (h TagCategory) Ensure(wanted *api.TagCategory) (err error) {
	for i := 0; i < 10; i++ {
		err = h.Create(wanted)
		if err == nil {
			return
		}
		found := false
		if errors.Is(err, &Conflict{}) {
			var cat *api.TagCategory
			cat, found, err = h.Find(wanted.Name)
			if found {
				*wanted = *cat
				break
			}
		}
	}
	return
}
