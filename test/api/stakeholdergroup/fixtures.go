package stakeholdergroup

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Set of valid resources for tests and reuse.
func Samples() (samples map[string]api.StakeholderGroup) {
	samples = map[string]api.StakeholderGroup{
		"Mgmt": {
			Name:        "Mgmt",
			Description: "Management stakeholder group.",
		},
		"Engineering": {
			Name:        "Engineering",
			Description: "Engineering team.",
		},
	}
	return
}

// Create a StakeholderGroup.
func Create(r *api.StakeholderGroup) (err error) {
	err = Client.Post(api.StakeholderGroupsRoot, &r)
	return
}

// Retrieve the StakeholderGroup.
func Get(r *api.StakeholderGroup) (err error) {
	err = Client.Get(client.Path(api.StakeholderGroupRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the StakeholderGroup.
func Update(r *api.StakeholderGroup) (err error) {
	err = Client.Put(client.Path(api.StakeholderGroupRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the StakeholderGroup.
func Delete(r *api.StakeholderGroup) (err error) {
	err = Client.Delete(client.Path(api.StakeholderGroupRoot, client.Params{api.ID: r.ID}))
	return
}

// List StakeholderGroups.
func List(r []*api.StakeholderGroup) (err error) {
	err = Client.Get(api.StakeholderGroupsRoot, &r)
	return
}
