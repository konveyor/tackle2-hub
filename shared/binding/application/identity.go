package application

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Identity sub-resource API.
type Identity struct {
	client *client.Client
	appId  uint
}

// List identities.
func (h Identity) List() (list []api.Identity, err error) {
	p := client.Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	path := client.Path(api.AppIdentitiesRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, &list, p)
	if err != nil {
		return
	}
	return
}

// Direct finds identities associated with the application.
func (h Identity) Direct(role string) (r *api.Identity, found bool, err error) {
	list := []api.Identity{}
	p := client.Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	filter := client.Filter{}
	filter.And("role").Eq(role)
	path := client.
		Path(api.AppIdentitiesRoute).
		Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, &list, p, filter.Param())
	if err != nil {
		return
	}
	for i := range list {
		r = &list[i]
		found = true
		return
	}
	return
}

// Indirect returns identities associated indirectly with the application.
func (h Identity) Indirect(kind string) (r *api.Identity, found bool, err error) {
	list := []api.Identity{}
	p := client.Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	filter := client.Filter{}
	filter.And("kind").Eq(kind)
	filter.And("default").Eq(true)
	err = h.client.Get(api.IdentitiesRoute, &list, p, filter.Param())
	if err != nil {
		return
	}
	for i := range list {
		r = &list[i]
		found = true
		return
	}
	return
}

// Search returns a search engine.
func (h Identity) Search() (s IdentitySearch) {
	s.api = &h
	return
}

// IdentitySearch engine.
type IdentitySearch struct {
	api        *Identity
	predicates []func() (*api.Identity, bool, error)
}

// Direct adds a direct search predicate.
func (q IdentitySearch) Direct(role string) IdentitySearch {
	q.predicates = append(
		q.predicates,
		func() (*api.Identity, bool, error) {
			return q.api.Direct(role)
		})
	return q
}

// Indirect adds an indirect search predicate.
func (q IdentitySearch) Indirect(kind string) IdentitySearch {
	q.predicates = append(
		q.predicates,
		func() (*api.Identity, bool, error) {
			return q.api.Indirect(kind)
		})
	return q
}

// Find performs the search.
func (q IdentitySearch) Find() (r *api.Identity, found bool, err error) {
	for _, p := range q.predicates {
		r, found, err = p()
		if err != nil || found {
			break
		}
	}
	return
}
