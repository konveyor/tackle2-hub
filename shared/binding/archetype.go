package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Archetype API.
type Archetype struct {
	client *Client
}

// Create a Archetype.
func (h *Archetype) Create(r *api.Archetype) (err error) {
	err = h.client.Post(api.ArchetypesRoute, &r)
	return
}

// Get a Archetype by ID.
func (h *Archetype) Get(id uint) (r *api.Archetype, err error) {
	r = &api.Archetype{}
	path := Path(api.ArchetypeRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Archetypes.
func (h *Archetype) List() (list []api.Archetype, err error) {
	list = []api.Archetype{}
	err = h.client.Get(api.ArchetypesRoute, &list)
	return
}

// Update a Archetype.
func (h *Archetype) Update(r *api.Archetype) (err error) {
	path := Path(api.ArchetypeRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Archetype.
func (h *Archetype) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ArchetypeRoute).Inject(Params{api.ID: id}))
	return
}
