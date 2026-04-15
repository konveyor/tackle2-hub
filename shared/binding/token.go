package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Token API.
type Token struct {
	client RestClient
}

// Create a RuleSet.
func (h Token) Create(r *api.TokenRequest) (err error) {
	err = h.client.Post(api.AuthTokensRoute, r)
	return
}

// Get a Token by ID.
func (h Token) Get(id uint) (r *api.Token, err error) {
	r = &api.Token{}
	path := Path(api.AuthTokenRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tokens.
func (h Token) List() (list []api.Token, err error) {
	list = []api.Token{}
	err = h.client.Get(api.AuthTokensRoute, &list)
	return
}

// Delete a Token.
func (h Token) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.AuthTokenRoute).Inject(Params{api.ID: id}))
	return
}
