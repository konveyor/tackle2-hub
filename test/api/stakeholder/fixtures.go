package stakeholder

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Set of valid resources for tests and reuse.
func Samples() (samples map[string]api.Stakeholder) {
	samples = map[string]api.Stakeholder{
		"Alice": {
			Name:  "Alice",
			Email: "alice@acme.local",
		},
		"Bob": {
			Name:  "Bob",
			Email: "bob@acme-supplier.local",
		},
	}
	return
}

//Samples2 := struct{
//	Alice api.StakeHolder
//}{
//	Alice: {
//		Name:  "Alice",
//		Email: "alice@acme.local",
//	},
//}

// Create a Stakeholder.
func Create(r *api.Stakeholder) (err error) {
	err = Client.Post(api.StakeholdersRoot, &r)
	return
}

// Retrieve the Stakeholder.
func Get(r *api.Stakeholder) (err error) {
	err = Client.Get(client.Path(api.StakeholderRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the Stakeholder.
func Update(r *api.Stakeholder) (err error) {
	err = Client.Put(client.Path(api.StakeholderRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the Stakeholder.
func Delete(r *api.Stakeholder) (err error) {
	err = Client.Delete(client.Path(api.StakeholderRoot, client.Params{api.ID: r.ID}))
	return
}

// List Stakeholders.
func List(r []*api.Stakeholder) (err error) {
	err = Client.Get(api.StakeholdersRoot, &r)
	return
}
