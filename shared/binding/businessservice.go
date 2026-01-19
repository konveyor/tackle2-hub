package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// BusinessService API.
type BusinessService struct {
	client *Client
}

// Create a BusinessService.
func (h BusinessService) Create(r *api.BusinessService) (err error) {
	err = h.client.Post(api.BusinessServicesRoute, r)
	return
}

// Get a BusinessService by ID.
func (h BusinessService) Get(id uint) (r *api.BusinessService, err error) {
	r = &api.BusinessService{}
	path := Path(api.BusinessServiceRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List BusinessServices.
func (h BusinessService) List() (list []api.BusinessService, err error) {
	list = []api.BusinessService{}
	err = h.client.Get(api.BusinessServicesRoute, &list)
	return
}

// Update a BusinessService.
func (h BusinessService) Update(r *api.BusinessService) (err error) {
	path := Path(api.BusinessServiceRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a BusinessService.
func (h BusinessService) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.BusinessServiceRoute).Inject(Params{api.ID: id}))
	return
}
