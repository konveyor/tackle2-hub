package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Tag API.
type Tag struct {
	// hub API client.
	client *Client
}

//
// Create a Tag.
func (h *Tag) Create(r *api.Tag) (err error) {
	err = h.client.Post(api.TagsRoot, &r)
	return
}

//
// Get a Tag by ID.
func (h *Tag) Get(id uint) (r *api.Tag, err error) {
	r = &api.Tag{}
	path := Path(api.TagRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Tags.
func (h *Tag) List() (list []api.Tag, err error) {
	list = []api.Tag{}
	err = h.client.Get(api.TagsRoot, &list)
	return
}

//
// Update a Tag.
func (h *Tag) Update(r *api.Tag) (err error) {
	path := Path(api.TagRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a Tag.
func (h *Tag) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TagRoot).Inject(Params{api.ID: id}))
	return
}
