package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Archetype API.
type Archetype struct {
	client *Client
}

// Create a Archetype.
func (h *Archetype) Create(r *api2.Archetype) (err error) {
	err = h.client.Post(api2.ArchetypesRoute, &r)
	return
}

// Get a Archetype by ID.
func (h *Archetype) Get(id uint) (r *api2.Archetype, err error) {
	r = &api2.Archetype{}
	path := Path(api2.ArchetypeRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Archetypes.
func (h *Archetype) List() (list []api2.Archetype, err error) {
	list = []api2.Archetype{}
	err = h.client.Get(api2.ArchetypesRoute, &list)
	return
}

// Update a Archetype.
func (h *Archetype) Update(r *api2.Archetype) (err error) {
	path := Path(api2.ArchetypeRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Archetype.
func (h *Archetype) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.ArchetypeRoute).Inject(Params{api2.ID: id}))
	return
}
