package addon

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
// Create a tag.
func (h *Tag) Create(r *api.Tag) (err error) {
	err = h.client.Post(api.TagsRoot, r)
	if err == nil {
		Log.Info(
			"Addon created: tag.",
			"object",
			r)
	}
	return
}

//
// Get a tag by ID.
func (h *Tag) Get(id uint) (r *api.Tag, err error) {
	r = &api.Tag{}
	path := Params{api.ID: id}.inject(api.TagRoot)
	err = h.client.Get(path, r)
	return
}

//
// List tags.
func (h *Tag) List() (list []api.Tag, err error) {
	list = []api.Tag{}
	err = h.client.Get(api.TagsRoot, &list)
	return
}

//
// Delete a tag.
func (h *Tag) Delete(r *api.Tag) (err error) {
	path := Params{api.ID: r.ID}.inject(api.TagRoot)
	err = h.client.Delete(path)
	if err == nil {
		Log.Info(
			"Addon deleted: tag.",
			"object",
			r)
	}
	return
}

//
// Find by name and type.
func (h *Tag) Find(name string, tp uint) (r *api.Tag, found bool, err error) {
	list := []api.Tag{}
	err = h.client.Get(api.TagsRoot, &list)
	if err != nil {
		return
	}
	for i := range list {
		if name == list[i].Name && tp == list[i].TagType.ID {
			r = &list[i]
			found = true
			break
		}
	}
	return
}

//
// Ensure a tag exists.
func (h *Tag) Ensure(wanted *api.Tag) (err error) {
	tag, found, err := h.Find(wanted.Name, wanted.TagType.ID)
	if err != nil {
		return
	}
	if !found {
		err = h.Create(wanted)
	} else {
		*wanted = *tag
	}
	return
}

//
// TagType API.
type TagType struct {
	// hub API client.
	client *Client
}

//
// Create a tag-type.
func (h *TagType) Create(m *api.TagType) (err error) {
	err = h.client.Post(api.TagTypesRoot, m)
	if err == nil {
		Log.Info(
			"Addon created: tag(type).",
			"object",
			m)
	}
	return
}

//
// Get a tag-type by ID.
func (h *TagType) Get(id uint) (r *api.TagType, err error) {
	r = &api.TagType{}
	path := Params{api.ID: id}.inject(api.TagTypeRoot)
	err = h.client.Get(path, r)
	return
}

//
// List tag-types.
func (h *TagType) List() (list []api.TagType, err error) {
	list = []api.TagType{}
	err = h.client.Get(api.TagTypesRoot, &list)
	return
}

//
// Delete a tag-type.
func (h *TagType) Delete(r *api.TagType) (err error) {
	path := Params{api.ID: r.ID}.inject(api.TagTypeRoot)
	err = h.client.Delete(path)
	if err == nil {
		Log.Info(
			"Addon deleted: tag(type).",
			"object",
			r)
	}
	return
}

//
// Find by name.
func (h *TagType) Find(name string) (r *api.TagType, found bool, err error) {
	list := []api.TagType{}
	err = h.client.Get(api.TagTypesRoot, &list)
	if err != nil {
		return
	}
	for i := range list {
		if name == list[i].Name {
			r = &list[i]
			found = true
			break
		}
	}
	return
}

//
// Ensure a tag-type exists.
func (h *TagType) Ensure(wanted *api.TagType) (err error) {
	tp, found, err := h.Find(wanted.Name)
	if err != nil {
		return
	}
	if !found {
		err = h.Create(wanted)
	} else {
		*wanted = *tp
	}
	return
}
