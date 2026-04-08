package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Permission API.
type Permission struct {
	client RestClient
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
