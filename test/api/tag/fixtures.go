package tag

import (
	"fmt"
	"testing"

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
func Create(t *testing.T, r *api.Tag) {
	err := Client.Post(api.TagsRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error())
	}
}

//
// Delete the Tag.
func Delete(t *testing.T, r *api.Tag) {
	err := Client.Delete(fmt.Sprintf("%s/%d", api.TagsRoot, r.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
