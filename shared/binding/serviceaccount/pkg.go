package serviceaccount

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// New creates a new ServiceAccount binding.
func New(c client.RestClient) (h ServiceAccount) {
	h = ServiceAccount{client: c}
	return
}

// ServiceAccount API.
type ServiceAccount struct {
	client client.RestClient
}

// Create a ServiceAccount.
func (h ServiceAccount) Create(r *api.ServiceAccount) (err error) {
	err = h.client.Post(api.ServiceAccountsRoute, r)
	return
}

// Get a ServiceAccount by ID.
func (h ServiceAccount) Get(id uint) (r *api.ServiceAccount, err error) {
	r = &api.ServiceAccount{}
	path := client.Path(api.ServiceAccountRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List ServiceAccounts.
func (h ServiceAccount) List() (list []api.ServiceAccount, err error) {
	list = []api.ServiceAccount{}
	err = h.client.Get(api.ServiceAccountsRoute, &list)
	return
}

// Update a ServiceAccount.
func (h ServiceAccount) Update(r *api.ServiceAccount) (err error) {
	path := client.Path(api.ServiceAccountRoute).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a ServiceAccount.
func (h ServiceAccount) Delete(id uint) (err error) {
	path := client.Path(api.ServiceAccountRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// Select returns the API for a selected service account.
func (h ServiceAccount) Select(id uint) (h2 Selected) {
	h2 = Selected{
		Token: Token{
			client: h.client,
			saId:   id,
		},
	}
	return
}

// Selected service account API.
type Selected struct {
	Token Token
}

// Token API for a selected service account.
type Token struct {
	client client.RestClient
	saId   uint
}

// Create an api-key token for the service account.
func (h Token) Create(r *api.PAT) (err error) {
	path := client.Path(api.ServiceAccountTokensRoute).Inject(client.Params{api.ID: h.saId})
	err = h.client.Post(path, r)
	return
}
