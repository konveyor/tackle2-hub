package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Dependency API.
type Dependency struct {
	// hub API client.
	client *Client
}

//
// Create a Dependency.
func (h *Dependency) Create(r *api.Dependency) (err error) {
	err = h.client.Post(api.DependenciesRoot, &r)
	return
}

//
// Get a Dependency by ID.
func (h *Dependency) Get(id uint) (r *api.Dependency, err error) {
	r = &api.Dependency{}
	path := Path(api.DependencyRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Dependencies.
func (h *Dependency) List() (list []api.Dependency, err error) {
	list = []api.Dependency{}
	err = h.client.Get(api.DependenciesRoot, &list)
	return
}

//
// Update a Dependency.
func (h *Dependency) Update(r *api.Dependency) (err error) {
	path := Path(api.DependencyRoot).Inject(Params{api.ID: r.From.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a Dependency.
func (h *Dependency) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.DependencyRoot).Inject(Params{api.ID: id}))
	return
}
