package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// IdpClient API.
type IdpClient struct {
	client client.RestClient
}

// Create a client.
func (h IdpClient) Create(r *api.IdpClient) (err error) {
	err = h.client.Post(api.IdpClientsRoute, r)
	return
}

// Get a client by ID.
func (h IdpClient) Get(id uint) (r *api.IdpClient, err error) {
	r = &api.IdpClient{}
	path := client.Path(api.IdpClientRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List clients.
func (h IdpClient) List() (list []api.IdpClient, err error) {
	list = []api.IdpClient{}
	err = h.client.Get(api.IdpClientsRoute, &list)
	return
}

// Update a client.
func (h IdpClient) Update(r *api.IdpClient) (err error) {
	path := client.Path(api.IdpClientRoute).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a client.
func (h IdpClient) Delete(id uint) (err error) {
	path := client.Path(api.IdpClientRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}
