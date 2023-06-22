package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// BusinessService API.
type BusinessService struct {
	// hub API client.
	Client *Client
}

//
// Create a BusinessService.
func (h *BusinessService) Create(r *api.BusinessService) (err error) {
	err = h.Client.Post(api.BusinessServicesRoot, &r)
	return
}

//
// Get a BusinessService by ID.
func (h *BusinessService) Get(id uint) (r *api.BusinessService, err error) {
	r = &api.BusinessService{}
	path := Path(api.BusinessServiceRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List BusinessServices.
func (h *BusinessService) List() (list []api.BusinessService, err error) {
	list = []api.BusinessService{}
	err = h.Client.Get(api.BusinessServicesRoot, &list)
	return
}

//
// Update a BusinessService.
func (h *BusinessService) Update(r *api.BusinessService) (err error) {
	path := Path(api.BusinessServiceRoot).Inject(Params{api.ID: r.ID})
	err = h.Client.Put(path, r)
	return
}

//
// Delete a BusinessService.
func (h *BusinessService) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.BusinessServiceRoot).Inject(Params{api.ID: id}))
	return
}
