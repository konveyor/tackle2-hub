package binding

import (
	"errors"

	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Tag API.
type Tag struct {
	client *Client
}

// Create a Tag.
func (h *Tag) Create(r *api2.Tag) (err error) {
	err = h.client.Post(api2.TagsRoute, &r)
	return
}

// Get a Tag by ID.
func (h *Tag) Get(id uint) (r *api2.Tag, err error) {
	r = &api2.Tag{}
	path := Path(api2.TagRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tags.
func (h *Tag) List() (list []api2.Tag, err error) {
	list = []api2.Tag{}
	err = h.client.Get(api2.TagsRoute, &list)
	return
}

// Update a Tag.
func (h *Tag) Update(r *api2.Tag) (err error) {
	path := Path(api2.TagRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Tag.
func (h *Tag) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.TagRoute).Inject(Params{api2.ID: id}))
	return
}

// Find by name and type.
func (h *Tag) Find(name string, category uint) (r *api2.Tag, found bool, err error) {
	list := []api2.Tag{}
	path := Path(api2.TagCategoryTagsRoute).Inject(Params{api2.ID: category})
	err = h.client.Get(
		path,
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

// Ensure a tag exists.
func (h *Tag) Ensure(wanted *api2.Tag) (err error) {
	for i := 0; i < 10; i++ {
		err = h.Create(wanted)
		if err == nil {
			return
		}
		found := false
		if errors.Is(err, &Conflict{}) {
			var tag *api2.Tag
			tag, found, err = h.Find(wanted.Name, wanted.Category.ID)
			if found {
				*wanted = *tag
				break
			}
		}
	}
	return
}
