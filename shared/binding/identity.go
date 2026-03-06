package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Identity API.
type Identity struct {
	client    RestClient
	decrypted bool
}

// Decrypted enables decryption.
// Returned resources with fields decrypted.
func (h Identity) Decrypted() (h2 Identity) {
	h2 = Identity{client: h.client, decrypted: true}
	return
}

// Create an Identity.
func (h Identity) Create(r *api.Identity) (err error) {
	err = h.client.Post(api.IdentitiesRoute, r)
	return
}

// Get an Identity by ID.
func (h Identity) Get(id uint) (r *api.Identity, err error) {
	r = &api.Identity{}
	p := h.params()
	path := Path(api.IdentityRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r, p...)
	return
}

// List Identities.
func (h Identity) List() (list []api.Identity, err error) {
	list = []api.Identity{}
	p := h.params()
	err = h.client.Get(api.IdentitiesRoute, &list, p...)
	return
}

// Find Identities.
func (h Identity) Find(filter Filter) (list []api.Identity, err error) {
	list = []api.Identity{}
	p := h.params(filter)
	err = h.client.Get(api.IdentitiesRoute, &list, p...)
	return
}

// Update an Identity.
func (h Identity) Update(r *api.Identity) (err error) {
	path := Path(api.IdentityRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete an Identity.
func (h Identity) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.IdentityRoute).Inject(Params{api.ID: id}))
	return
}

// params returns parameters.
func (h Identity) params(filter ...Filter) (param []Param) {
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
