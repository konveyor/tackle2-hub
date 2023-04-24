package tag

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Create a Tag.
func Create(r *api.Tag) (err error) {
	err = Client.Post(api.TagsRoot, &r)
	return
}

// Retrieve the Tag.
func Get(r *api.Tag) (err error) {
	err = Client.Get(client.Path(api.TagRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the Tag.
func Update(r *api.Tag) (err error) {
	err = Client.Put(client.Path(api.TagRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the Tag.
func Delete(r *api.Tag) (err error) {
	err = Client.Delete(client.Path(api.TagRoot, client.Params{api.ID: r.ID}))
	return
}

// List Tags.
func List(r []*api.Tag) (err error) {
	err = Client.Get(api.TagsRoot, &r)
	return
}
