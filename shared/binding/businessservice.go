package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// BusinessService API.
type BusinessService struct {
	client *Client
}

// Create a BusinessService.
func (h *BusinessService) Create(r *api2.BusinessService) (err error) {
	err = h.client.Post(api2.BusinessServicesRoute, &r)
	return
}

// Get a BusinessService by ID.
func (h *BusinessService) Get(id uint) (r *api2.BusinessService, err error) {
	r = &api2.BusinessService{}
	path := Path(api2.BusinessServiceRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List BusinessServices.
func (h *BusinessService) List() (list []api2.BusinessService, err error) {
	list = []api2.BusinessService{}
	err = h.client.Get(api2.BusinessServicesRoute, &list)
	return
}

// Update a BusinessService.
func (h *BusinessService) Update(r *api2.BusinessService) (err error) {
	path := Path(api2.BusinessServiceRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a BusinessService.
func (h *BusinessService) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.BusinessServiceRoute).Inject(Params{api2.ID: id}))
	return
}
