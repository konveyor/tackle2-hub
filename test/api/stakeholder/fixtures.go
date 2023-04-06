package stakeholder

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
func Samples() (samples []api.Stakeholder) {
	samples = []api.Stakeholder{
		{
			Name:  "Alice",
			Email: "alice@acme.local",
		},
		{
			Name:  "Bob",
			Email: "bob@acme-supplier.local",
		},
	}
	return
}

//
// Create a Stakeholder.
func Create(r *api.Stakeholder) (err error) {
	err = Client.Post(api.StakeholdersRoot, &r)
	return
}

//
// Retrieve the Stakeholder.
func Get(r *api.Stakeholder) (err error) {
	err = Client.Get(c.Path(api.StakeholderRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Update the Stakeholder.
func Update(r *api.Stakeholder) (err error) {
	err = Client.Put(c.Path(api.StakeholderRoot, c.Params{api.ID: r.ID}), &r)
	return
}

//
// Delete the Stakeholder.
func Delete(r *api.Stakeholder) (err error) {
	err = Client.Delete(c.Path(api.StakeholderRoot, c.Params{api.ID: r.ID}))
	return
}

//
// List Stakeholders.
func List(r []*api.Stakeholder) (err error) {
	err = Client.Get(api.StakeholdersRoot, &r)
	return
}
