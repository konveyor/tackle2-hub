package archetype

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) (h Archetype) {
	h = Archetype{client: client}
	return
}

// Archetype API.
type Archetype struct {
	client client.RestClient
}

// Create a Archetype.
func (h Archetype) Create(r *api.Archetype) (err error) {
	err = h.client.Post(api.ArchetypesRoute, r)
	return
}

// Get a Archetype by ID.
func (h Archetype) Get(id uint) (r *api.Archetype, err error) {
	r = &api.Archetype{}
	path := client.Path(api.ArchetypeRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Archetypes.
func (h Archetype) List() (list []api.Archetype, err error) {
	list = []api.Archetype{}
	err = h.client.Get(api.ArchetypesRoute, &list)
	return
}

// Update a Archetype.
func (h Archetype) Update(r *api.Archetype) (err error) {
	path := client.Path(api.ArchetypeRoute).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Archetype.
func (h Archetype) Delete(id uint) (err error) {
	path := client.Path(api.ArchetypeRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// Select returns the API for a selected archetype.
func (h Archetype) Select(id uint) (h2 Selected) {
	h2 = Selected{}
	h2.Assessment = Assessment{client: h.client, archetypeId: id}
	return
}

// Selected archetype API.
type Selected struct {
	Assessment Assessment
}
