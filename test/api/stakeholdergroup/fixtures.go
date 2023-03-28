package stakeholdergroup

import (
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
func Samples() (samples []api.StakeholderGroup) {
	samples = []api.StakeholderGroup{
		{
			Name:        "Mgmt",
			Description: "Management stakeholder group.",
		},
		{
			Name:        "Engineering",
			Description: "Engineering team.",
		},
	}
	return
}

//
// Create.
func Create(t *testing.T, r *api.StakeholderGroup) {
	err := Client.Post(api.StakeholderGroupsRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error())
	}
}

//
// Delete.
func Delete(t *testing.T, r *api.StakeholderGroup) {
	err := Client.Delete(client.Path(api.StakeholderGroupRoot, client.Params{api.ID: r.ID}))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
