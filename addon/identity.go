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
	p := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	path := Params{api.ID: id}.inject(api.IdentityRoot)
	err = h.client.Get(path, r, p)
	if err != nil {
		return
	}
	m := r.Model()
	r.With(m)
	return
}

//
// List identities.
func (h *Identity) List() (list []api.Identity, err error) {
	list = []api.Identity{}
	p := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	err = h.client.Get(api.IdentitiesRoot, &list, p)
	if err != nil {
		return
	}
	for i := range list {
		r := &list[i]
		m := r.Model()
		r.With(m)
	}
	return
}
