package tag

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = client.Client
)

//
// Set of valid resources for tests and reuse.
func Samples() (samples []api.Tag) {
	samples = []api.Tag{
		{
			Name: "Test Linux",
			Category: api.Ref{
				ID: 1, // Category from seeds.
			},
		},
		{
			Name: "Test RHEL",
			Category: api.Ref{
				ID: 2, // Category from seeds.
			},
		},
	}

	return
}

//
// Create a Tag.
func Create(r *api.Tag) (err error) {
	err = Client.Post(api.TagsRoot, &r)
	return
}

//
// Retrieve the Tag.
func Get(r *api.Tag) (err error) {
	err = Client.Get(client.Path(api.TagRoot, client.Params{api.ID: r.ID}), &r)
	return
}

//
// Update the Tag.
func Update(r *api.Tag) (err error) {
	err = Client.Put(client.Path(api.TagRoot, client.Params{api.ID: r.ID}), &r)
	return
}

//
// Delete the Tag.
func Delete(r *api.Tag) (err error) {
	err = Client.Delete(client.Path(api.TagRoot, client.Params{api.ID: r.ID}))
	return
}

//
// List Tags.
func List(r []*api.Tag) (err error) {
	err = Client.Get(api.TagsRoot, &r)
	return
}
