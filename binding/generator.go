package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Generator API.
type Generator struct {
	client *Client
}

// Create a Generator.
func (h *Generator) Create(r *api2.Generator) (err error) {
	err = h.client.Post(api2.GeneratorsRoute, &r)
	return
}

// Get a Generator by ID.
func (h *Generator) Get(id uint) (r *api2.Generator, err error) {
	r = &api2.Generator{}
	path := Path(api2.GeneratorRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Generators.
func (h *Generator) List() (list []api2.Generator, err error) {
	list = []api2.Generator{}
	err = h.client.Get(api2.GeneratorsRoute, &list)
	return
}

// Update a Generator.
func (h *Generator) Update(r *api2.Generator) (err error) {
	path := Path(api2.GeneratorRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Generator.
func (h *Generator) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.GeneratorRoute).Inject(Params{api2.ID: id}))
	return
}
