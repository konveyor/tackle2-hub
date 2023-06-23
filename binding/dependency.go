package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Dependency API.
type Dependency struct {
	// hub API client.
	Client *Client
}

//
// Create a Dependency.
func (h *Dependency) Create(r *api.Dependency) (err error) {
	err = h.Client.Post(api.DependenciesRoot, &r)
	return
}

//
// Get a Dependency by ID.
func (h *Dependency) Get(id uint) (r *api.Dependency, err error) {
	r = &api.Dependency{}
	path := Path(api.DependencyRoot).Inject(Params{api.ID: id})
	err = h.Client.Get(path, r)
	return
}

//
// List Dependencies.
func (h *Dependency) List() (list []api.Dependency, err error) {
	list = []api.Dependency{}
	err = h.Client.Get(api.DependenciesRoot, &list)
	return
}

//
// Delete a Dependency.
func (h *Dependency) Delete(id uint) (err error) {
	err = h.Client.Delete(Path(api.DependencyRoot).Inject(Params{api.ID: id}))
	return
}
