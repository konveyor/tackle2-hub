package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Dependency API.
type Dependency struct {
	// hub API client.
	client *Client
}

// Create a Dependency.
func (h *Dependency) Create(r *api2.Dependency) (err error) {
	err = h.client.Post(api2.DependenciesRoute, &r)
	return
}

// Get a Dependency by ID.
func (h *Dependency) Get(id uint) (r *api2.Dependency, err error) {
	r = &api2.Dependency{}
	path := Path(api2.DependencyRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Dependencies.
func (h *Dependency) List() (list []api2.Dependency, err error) {
	list = []api2.Dependency{}
	err = h.client.Get(api2.DependenciesRoute, &list)
	return
}

// Delete a Dependency.
func (h *Dependency) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.DependencyRoute).Inject(Params{api2.ID: id}))
	return
}
