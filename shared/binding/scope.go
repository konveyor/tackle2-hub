package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Scope API.
type Scope struct {
	client RestClient
}

// List Scopes.
func (h Scope) List() (list []api.Scope, err error) {
	list = []api.Scope{}
	err = h.client.Get(api.AuthScopesRoute, &list)
	return
}

// Find Scopes.
func (h Scope) Find(filter Filter) (list []api.Scope, err error) {
	list = []api.Scope{}
	err = h.client.Get(api.AuthScopesRoute, &list, filter.Param())
	return
}
