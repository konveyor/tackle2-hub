package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Identity API.
type Identity struct {
	client *Client
}

// Create a Identity.
func (h *Identity) Create(r *api2.Identity) (err error) {
	err = h.client.Post(api2.IdentitiesRoute, &r)
	return
}

// Get a decrypted Identity by ID.
func (h *Identity) Get(id uint) (r *api2.Identity, err error) {
	r = &api2.Identity{}
	p := Param{
		Key:   api2.Decrypted,
		Value: "1",
	}
	path := Path(api2.IdentityRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r, p)
	return
}

// List decrypted Identities.
func (h *Identity) List() (list []api2.Identity, err error) {
	list = []api2.Identity{}
	p := Param{
		Key:   api2.Decrypted,
		Value: "1",
	}
	err = h.client.Get(api2.IdentitiesRoute, &list, p)
	return
}

// Find decrypted Identities.
func (h *Identity) Find(filter Filter) (list []api2.Identity, err error) {
	list = []api2.Identity{}
	p := Param{
		Key:   api2.Decrypted,
		Value: "1",
	}
	err = h.client.Get(api2.IdentitiesRoute, &list, p, filter.Param())
	return
}

// Update a Identity.
func (h *Identity) Update(r *api2.Identity) (err error) {
	path := Path(api2.IdentityRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Identity.
func (h *Identity) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.IdentityRoute).Inject(Params{api2.ID: id}))
	return
}
