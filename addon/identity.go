package addon

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Identity API.
type Identity struct {
	// hub API client.
	client *Client
}

//
// Get an identity by ID.
func (h *Identity) Get(id uint) (r *api.Identity, err error) {
	r = &api.Identity{}
	path := Params{api.ID: id}.inject(api.IdentityRoot)
	err = h.client.Get(path, r)
	if err != nil {
		return
	}
	m := r.Model()
	err = m.Decrypt(Addon.secret.Hub.Encryption.Passphrase)
	r.With(m)
	return
}

//
// List identities.
func (h *Identity) List() (list []api.Identity, err error) {
	list = []api.Identity{}
	err = h.client.Get(api.IdentitiesRoot, &list)
	if err != nil {
		return
	}
	for i := range list {
		r := &list[i]
		m := r.Model()
		err = m.Decrypt(Addon.secret.Hub.Encryption.Passphrase)
		r.With(m)
		if err != nil {
			return
		}
	}
	return
}
