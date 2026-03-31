package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// User API.
type User struct {
	client RestClient
}

// Create a User.
func (h User) Create(r *api.User) (err error) {
	err = h.client.Post(api.UsersRoute, r)
	return
}

// Get a User by ID.
func (h User) Get(id uint) (r *api.User, err error) {
	r = &api.User{}
	path := Path(api.UserRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Users.
func (h User) List() (list []api.User, err error) {
	list = []api.User{}
	err = h.client.Get(api.UsersRoute, &list)
	return
}

// Find Users.
func (h User) Find(filter Filter) (list []api.User, err error) {
	list = []api.User{}
	err = h.client.Get(api.UsersRoute, &list, filter.Param())
	return
}

// Update a User.
func (h User) Update(r *api.User) (err error) {
	path := Path(api.UserRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a User.
func (h User) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.UserRoute).Inject(Params{api.ID: id}))
	return
}
