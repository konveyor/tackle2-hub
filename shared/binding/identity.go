package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Identity API.
type Identity struct {
	client *Client
}

// Create a Identity.
func (h Identity) Create(r *api.Identity) (err error) {
	err = h.client.Post(api.IdentitiesRoute, r)
	return
}

// Get a decrypted Identity by ID.
func (h Identity) Get(id uint) (r *api.Identity, err error) {
	r = &api.Identity{}
	p := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	path := Path(api.IdentityRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r, p)
	return
}

// List decrypted Identities.
func (h Identity) List() (list []api.Identity, err error) {
	list = []api.Identity{}
	p := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	err = h.client.Get(api.IdentitiesRoute, &list, p)
	return
}

// Find decrypted Identities.
func (h Identity) Find(filter Filter) (list []api.Identity, err error) {
	list = []api.Identity{}
	p := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	err = h.client.Get(api.IdentitiesRoute, &list, p, filter.Param())
	return
}

// Update a Identity.
func (h Identity) Update(r *api.Identity) (err error) {
	path := Path(api.IdentityRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Identity.
func (h Identity) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.IdentityRoute).Inject(Params{api.ID: id}))
	return
}
