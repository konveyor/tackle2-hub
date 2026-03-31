package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// User API.
type User struct {
	client    RestClient
	decrypted bool
}

// Decrypted enables decryption.
// Returned resources with fields decrypted.
func (h User) Decrypted() (h2 User) {
	h2 = User{client: h.client, decrypted: true}
	return
}

// Create a User.
func (h User) Create(r *api.User) (err error) {
	err = h.client.Post(api.UsersRoute, r)
	return
}

// Get a User by ID.
func (h User) Get(id uint) (r *api.User, err error) {
	r = &api.User{}
	p := h.params()
	path := Path(api.UserRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r, p...)
	return
}

// List Users.
func (h User) List() (list []api.User, err error) {
	list = []api.User{}
	p := h.params()
	err = h.client.Get(api.UsersRoute, &list, p...)
	return
}

// Find Users.
func (h User) Find(filter Filter) (list []api.User, err error) {
	list = []api.User{}
	p := h.params(filter)
	err = h.client.Get(api.UsersRoute, &list, p...)
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

// params returns parameters.
func (h User) params(filter ...Filter) (param []Param) {
	if h.decrypted {
		param = append(
			param, Param{
				Key:   api.Decrypted,
				Value: "1",
			})
	}
	for _, f := range filter {
		param = append(param, f.Param())
	}
	return
}
