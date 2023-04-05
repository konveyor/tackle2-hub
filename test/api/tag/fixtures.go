package tag

import (
	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = c.Client
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
// Delete the Tag.
func Delete(r *api.Tag) (err error) {
	err = Client.Delete(c.Path(api.TagRoot, c.Params{api.ID: r.ID}))
	return
}
