package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Generator API.
type Generator struct {
	client *Client
}

// Create a Generator.
func (h Generator) Create(r *api.Generator) (err error) {
	err = h.client.Post(api.GeneratorsRoute, r)
	return
}

// Get a Generator by ID.
func (h Generator) Get(id uint) (r *api.Generator, err error) {
	r = &api.Generator{}
	path := Path(api.GeneratorRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Generators.
func (h Generator) List() (list []api.Generator, err error) {
	list = []api.Generator{}
	err = h.client.Get(api.GeneratorsRoute, &list)
	return
}

// Update a Generator.
func (h Generator) Update(r *api.Generator) (err error) {
	path := Path(api.GeneratorRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Generator.
func (h Generator) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.GeneratorRoute).Inject(Params{api.ID: id}))
	return
}
