package application

import (
	"errors"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Tag sub-resource API.
// Provides association management of tags to applications by name.
type Tag struct {
	client *client.Client
	appId  uint
	source *string
}

// Source sets the source for other operations on the associated tags.
func (h Tag) Source(name string) (h2 Tag) {
	h2.client = h.client
	h2.appId = h.appId
	h2.source = &name
	return
}

// Replace the associated tags for the source with a new set.
// Returns an error if the source is not set.
func (h Tag) Replace(ids []uint) (err error) {
	if h.source == nil {
		err = liberr.New("Source required.")
		return
	}
	path := client.Path(api.ApplicationTagsRoute).Inject(client.Params{api.ID: h.appId})
	query := []client.Param{}
	if h.source != nil {
		query = append(
			query,
			client.Param{
				Key:   api.Source,
				Value: *h.source,
			})
	}
	tags := []api.TagRef{}
	for _, id := range h.unique(ids) {
		tags = append(tags, api.TagRef{ID: id})
	}
	err = h.client.Put(path, tags, query...)
	return
}

// List associated tags.
// Returns a list of tag names.
func (h Tag) List() (list []api.TagRef, err error) {
	list = []api.TagRef{}
	path := client.
		Path(api.ApplicationTagsRoute).
		Inject(client.Params{api.ID: h.appId})
	query := []client.Param{}
	if h.source != nil {
		query = append(
			query,
			client.Param{
				Key:   api.Source,
				Value: *h.source,
			})
	}
	err = h.client.Get(path, &list, query...)
	return
}

// Add associates a tag with the application.
func (h Tag) Add(id uint) (err error) {
	path := client.Path(api.ApplicationTagsRoute).Inject(client.Params{api.ID: h.appId})
	tag := api.TagRef{ID: id}
	if h.source != nil {
		tag.Source = *h.source
	}
	err = h.client.Post(path, &tag)
	return
}

// Ensure ensures tag is associated with the application.
func (h Tag) Ensure(id uint) (err error) {
	err = h.Add(id)
	if errors.Is(err, &client.Conflict{}) {
		err = nil
	}
	return
}

// Delete ensures the tag is not associated with the application.
func (h Tag) Delete(id uint) (err error) {
	path := client.Path(
		api.ApplicationTagRoute).Inject(
		client.Params{
			api.ID:  h.appId,
			api.ID2: id})
	query := []client.Param{}
	if h.source != nil {
		query = append(
			query, client.Param{
				Key:   api.Source,
				Value: *h.source,
			})
	}
	err = h.client.Delete(path, query...)
	return
}

// unique ensures unique ids.
func (h Tag) unique(ids []uint) (u []uint) {
	mp := map[uint]int{}
	for _, id := range ids {
		mp[id] = 0
	}
	for id := range mp {
		u = append(u, id)
	}
	return
}
