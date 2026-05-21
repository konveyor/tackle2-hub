package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Role API.
type Role struct {
	client RestClient
}

// Create a Role.
func (h Role) Create(r *api.Role) (err error) {
	err = h.client.Post(api.RolesRoute, r)
	return
}

// Get a Role by ID.
func (h Role) Get(id uint) (r *api.Role, err error) {
	r = &api.Role{}
	path := Path(api.RoleRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Roles.
func (h Role) List() (list []api.Role, err error) {
	list = []api.Role{}
	err = h.client.Get(api.RolesRoute, &list)
	return
}

// Find Roles.
func (h Role) Find(filter Filter) (list []api.Role, err error) {
	list = []api.Role{}
	err = h.client.Get(api.RolesRoute, &list, filter.Param())
	return
}

// Update a Role.
func (h Role) Update(r *api.Role) (err error) {
	path := Path(api.RoleRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Role.
func (h Role) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.RoleRoute).Inject(Params{api.ID: id}))
	return
}
