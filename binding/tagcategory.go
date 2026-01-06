package binding

import (
	"errors"

	api2 "github.com/konveyor/tackle2-hub/api"
)

// TagCategory API.
type TagCategory struct {
	client *Client
}

// Create a TagCategory.
func (h *TagCategory) Create(r *api2.TagCategory) (err error) {
	err = h.client.Post(api2.TagCategoriesRoute, &r)
	return
}

// Get a TagCategory by ID.
func (h *TagCategory) Get(id uint) (r *api2.TagCategory, err error) {
	r = &api2.TagCategory{}
	path := Path(api2.TagCategoryRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List TagCategories.
func (h *TagCategory) List() (list []api2.TagCategory, err error) {
	list = []api2.TagCategory{}
	err = h.client.Get(api2.TagCategoriesRoute, &list)
	return
}

// Update a TagCategory.
func (h *TagCategory) Update(r *api2.TagCategory) (err error) {
	path := Path(api2.TagCategoryRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a TagCategory.
func (h *TagCategory) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.TagCategoryRoute).Inject(Params{api2.ID: id}))
	return
}

// Find by name.
func (h *TagCategory) Find(name string) (r *api2.TagCategory, found bool, err error) {
	list := []api2.TagCategory{}
	err = h.client.Get(
		api2.TagCategoriesRoute,
		&list,
		Param{
			Key:   api2.Name,
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
func (h *TagCategory) Ensure(wanted *api2.TagCategory) (err error) {
	for i := 0; i < 10; i++ {
		err = h.Create(wanted)
		if err == nil {
			return
		}
		found := false
		if errors.Is(err, &Conflict{}) {
			var cat *api2.TagCategory
			cat, found, err = h.Find(wanted.Name)
			if found {
				*wanted = *cat
				break
			}
		}
	}
	return
}
