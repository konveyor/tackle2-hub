package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Permission API.
type Permission struct {
	client RestClient
}

// Create a Permission.
func (h Permission) Create(r *api.Permission) (err error) {
	err = h.client.Post(api.PermissionsRoute, r)
	return
}

// Get a Permission by ID.
func (h Permission) Get(id uint) (r *api.Permission, err error) {
	r = &api.Permission{}
	path := Path(api.PermissionRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Permissions.
func (h Permission) List() (list []api.Permission, err error) {
	list = []api.Permission{}
	err = h.client.Get(api.PermissionsRoute, &list)
	return
}

// Find Permissions.
func (h Permission) Find(filter Filter) (list []api.Permission, err error) {
	list = []api.Permission{}
	err = h.client.Get(api.PermissionsRoute, &list, filter.Param())
	return
}

// Update a Permission.
func (h Permission) Update(r *api.Permission) (err error) {
	path := Path(api.PermissionRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Permission.
func (h Permission) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.PermissionRoute).Inject(Params{api.ID: id}))
	return
}
