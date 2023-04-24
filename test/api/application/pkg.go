package application

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

// Setup Hub API client
var Client = client.Client

// Create a Application.
func Create(r *api.Application) (err error) {
	err = Client.Post(api.ApplicationsRoot, &r)
	return
}

// Retrieve the Application.
func Get(r *api.Application) (err error) {
	err = Client.Get(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Update the Application.
func Update(r *api.Application) (err error) {
	err = Client.Put(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}), &r)
	return
}

// Delete the Application.
func Delete(r *api.Application) (err error) {
	err = Client.Delete(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}))
	return
}

// List Applications.
func List(r []*api.Application) (err error) {
	err = Client.Get(api.ApplicationsRoot, &r)
	return
}
