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
	path := Path(api.TagRoot).Inject(Params{api.ID: id})
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
	path := Path(api.TagRoot).Inject(Params{api.ID: r.ID})
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
func (h *Tag) Find(name string, category uint) (r *api.Tag, found bool, err error) {
	list := []api.Tag{}
	path := Path(api.TagCategoryTagsRoot).Inject(Params{api.ID: category})
	err = h.client.Get(path, &list, Param{Key: api.Name, Value: name})
	if err != nil {
		return
	}
	if len(list) > 0 {
		found = true
		r = &list[0]
	}
	return
}

//
// Ensure a tag exists.
func (h *Tag) Ensure(wanted *api.Tag) (err error) {
	tag, found, err := h.Find(wanted.Name, wanted.Category.ID)
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
// TagCategory API.
type TagCategory struct {
	// hub API client.
	client *Client
}

//
// Create a tag-type.
func (h *TagCategory) Create(m *api.TagCategory) (err error) {
	err = h.client.Post(api.TagCategoriesRoot, m)
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
func (h *TagCategory) Get(id uint) (r *api.TagCategory, err error) {
	r = &api.TagCategory{}
	path := Path(api.TagCategoryRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List tag-types.
func (h *TagCategory) List() (list []api.TagCategory, err error) {
	list = []api.TagCategory{}
	err = h.client.Get(api.TagCategoriesRoot, &list)
	return
}

//
// Delete a tag-type.
func (h *TagCategory) Delete(r *api.TagCategory) (err error) {
	path := Path(api.TagCategoryRoot).Inject(Params{api.ID: r.ID})
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
func (h *TagCategory) Find(name string) (r *api.TagCategory, found bool, err error) {
	list := []api.TagCategory{}
	err = h.client.Get(api.TagCategoriesRoot, &list, Param{Key: api.Name, Value: name})
	if err != nil {
		return
	}
	if len(list) > 0 {
		found = true
		r = &list[0]
	}
	return
}

//
// Ensure a tag-type exists.
func (h *TagCategory) Ensure(wanted *api.TagCategory) (err error) {
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
