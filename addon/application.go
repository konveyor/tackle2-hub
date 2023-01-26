package addon

import (
	"github.com/konveyor/tackle2-hub/api"
	"strconv"
)

//
// Application API.
type Application struct {
	// hub API client.
	client *Client
}

//
// Get an application by ID.
func (h *Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := Params{api.ID: id}.inject(api.ApplicationRoot)
	err = h.client.Get(path, r)
	return
}

//
// List applications.
func (h *Application) List() (list []api.Application, err error) {
	list = []api.Application{}
	err = h.client.Get(api.ApplicationsRoot, &list)
	return
}

//
// Update an application by ID.
func (h *Application) Update(r *api.Application) (err error) {
	path := Params{api.ID: r.ID}.inject(api.ApplicationRoot)
	err = h.client.Put(path, r)
	if err == nil {
		Log.Info(
			"Addon updated: application.",
			"id",
			r.ID)
	}
	return
}

//
// FindIdentity by kind.
func (h *Application) FindIdentity(id uint, kind string) (r *api.Identity, found bool, err error) {
	list := []api.Identity{}
	p1 := Param{
		Key:   api.AppId,
		Value: strconv.Itoa(int(id)),
	}
	p2 := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	path := Params{api.ID: id}.inject(api.IdentitiesRoot)
	err = h.client.Get(path, &list, p1, p2)
	if err != nil {
		return
	}
	for i := range list {
		r = &list[i]
		if r.Kind == kind {
			m := r.Model()
			r.With(m)
			found = true
			break
		}
	}
	return
}

//
// Bucket returns the bucket API.
func (h *Application) Bucket(id uint) (b *Bucket) {
	params := Params{
		api.Wildcard: "",
		api.ID:       id,
	}
	path := params.inject(api.AppBucketRoot)
	b = &Bucket{
		path:   path,
		client: h.client,
	}
	return
}

//
// Tags returns the tags API.
func (h *Application) Tags(id uint) (tg AppTags) {
	tg = AppTags{
		client: h.client,
		appId:  id,
	}
	return
}

//
// AppTags sub-resource API.
// Provides association management of tags to applications by name.
type AppTags struct {
	client *Client
	appId  uint
}

//
// List associated tags.
// Returns a list of tag names.
func (h *AppTags) List() (list []api.Ref, err error) {
	list = []api.Ref{}
	path := Params{api.ID: h.appId}.inject(api.ApplicationTagsRoot)
	err = h.client.Get(path, &list)
	return
}

//
// Add ensures tag is associated with the application.
func (h *AppTags) Add(id uint) (err error) {
	path := Params{api.ID: h.appId}.inject(api.ApplicationTagsRoot)
	err = h.client.Post(path, &api.Ref{ID: id})
	return
}

//
// Delete ensures the tag is not associated with the application.
func (h *AppTags) Delete(id uint) (err error) {
	path := Params{
		api.ID:  h.appId,
		api.ID2: id}.inject(api.ApplicationTagRoot)
	err = h.client.Delete(path)
	return
}
