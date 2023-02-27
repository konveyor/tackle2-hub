package addon

import (
	liberr "github.com/konveyor/controller/pkg/error"
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
	path := params.inject(api.AppBucketContentRoot)
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
	source *string
}

//
// Source sets the source for other operations on the associated tags.
func (h *AppTags) Source(name string) {
	h.source = &name
}

//
// Replace the associated tags for the source with a new set.
// Returns an error if the source is not set.
func (h *AppTags) Replace(ids []uint) (err error) {
	if h.source == nil {
		err = liberr.New("`source` must be set")
		return
	}
	path := Params{api.ID: h.appId}.inject(api.ApplicationTagsRoot)
	query := []Param{}
	if h.source != nil {
		query = append(query, Param{Key: api.Source, Value: *h.source})
	}

	tags := []api.TagRef{}
	for _, id := range ids {
		tags = append(tags, api.TagRef{ID: id})
	}

	err = h.client.Put(path, tags, query...)
	return
}

//
// List associated tags.
// Returns a list of tag names.
func (h *AppTags) List() (list []api.TagRef, err error) {
	list = []api.TagRef{}
	path := Params{api.ID: h.appId}.inject(api.ApplicationTagsRoot)
	query := []Param{}
	if h.source != nil {
		query = append(query, Param{Key: api.Source, Value: *h.source})
	}
	err = h.client.Get(path, &list, query...)
	return
}

//
// Add ensures tag is associated with the application.
func (h *AppTags) Add(id uint) (err error) {
	path := Params{api.ID: h.appId}.inject(api.ApplicationTagsRoot)

	tag := api.TagRef{ID: id}
	if h.source != nil {
		tag.Source = *h.source
	}
	err = h.client.Post(path, &tag)
	return
}

//
// Delete ensures the tag is not associated with the application.
func (h *AppTags) Delete(id uint) (err error) {
	path := Params{
		api.ID:  h.appId,
		api.ID2: id}.inject(api.ApplicationTagRoot)
	query := []Param{}
	if h.source != nil {
		query = append(query, Param{Key: api.Source, Value: *h.source})
	}
	err = h.client.Delete(path, query...)
	return
}

//
// Facts returns the tags API.
func (h *Application) Facts(id uint) (f AppFacts) {
	f = AppFacts{
		client: h.client,
		appId:  id,
	}
	return
}

//
// AppFacts sub-resource API.
// Provides association management of facts.
type AppFacts struct {
	client *Client
	appId  uint
}

//
// List associated tags.
// Returns a list of tag names.
func (h *AppFacts) List() (list []api.Fact, err error) {
	list = []api.Fact{}
	path := Params{api.ID: h.appId}.inject(api.ApplicationFactsRoot)
	err = h.client.Get(path, &list)
	return
}

//
// Get a fact.
func (h *AppFacts) Get(key string, valuePtr interface{}) (err error) {
	path := Params{
		api.ID:  h.appId,
		api.Key: key,
	}.inject(api.ApplicationFactRoot)
	err = h.client.Get(path, valuePtr)
	return
}

//
// Set a fact (created as needed).
func (h *AppFacts) Set(key string, value interface{}) (err error) {
	path := Params{
		api.ID:  h.appId,
		api.Key: key,
	}.inject(api.ApplicationFactRoot)
	err = h.client.Put(path, value)
	return
}

//
// Delete a fact.
func (h *AppFacts) Delete(key string) (err error) {
	path := Params{
		api.ID:  h.appId,
		api.Key: key,
	}.inject(api.ApplicationFactRoot)
	err = h.client.Delete(path)
	return
}
