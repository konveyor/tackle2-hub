package application

import (
	pathlib "path"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Fact sub-resource API.
// Provides association management of facts.
type Fact struct {
	client *client.Client
	appId  uint
	source string
}

// Source sets the source for other operations on the facts.
func (h Fact) Source(name string) (h2 Fact) {
	h2.client = h.client
	h2.appId = h.appId
	h2.source = name
	return
}

// Create a fact.
func (h Fact) Create(r *api.Fact) (err error) {
	params := client.Params{
		api.ID: h.appId,
	}
	path := client.Path(api.ApplicationFactsRoute).Inject(params)
	err = h.client.Post(path, r)
	return
}

// List facts.
func (h Fact) List() (facts api.Map, err error) {
	facts = api.Map{}
	key := api.FactKey("")
	key.Qualify(h.source)
	path := client.Path(api.ApplicationFactsRoute).Inject(
		client.Params{
			api.ID: h.appId,
		})
	path = pathlib.Join(path, string(key))
	err = h.client.Get(path, &facts)
	return
}

// Get a fact.
func (h Fact) Get(name string, value any) (err error) {
	key := api.FactKey(name)
	key.Qualify(h.source)
	path := client.Path(api.ApplicationFactRoute).Inject(
		client.Params{
			api.ID:  h.appId,
			api.Key: key,
		})
	err = h.client.Get(path, value)
	return
}

// Set a fact (created as needed).
func (h Fact) Set(name string, value any) (err error) {
	key := api.FactKey(name)
	key.Qualify(h.source)
	path := client.Path(api.ApplicationFactRoute).Inject(
		client.Params{
			api.ID:  h.appId,
			api.Key: key,
		})
	err = h.client.Put(path, value)
	return
}

// Delete a fact.
func (h Fact) Delete(name string) (err error) {
	key := api.FactKey(name)
	key.Qualify(h.source)
	path := client.Path(api.ApplicationFactRoute).Inject(
		client.Params{
			api.ID:  h.appId,
			api.Key: key,
		})
	err = h.client.Delete(path)
	return
}

// Replace facts.
func (h Fact) Replace(facts api.Map) (err error) {
	key := api.FactKey("")
	key.Qualify(h.source)
	path := client.Path(api.ApplicationFactsRoute).Inject(
		client.Params{
			api.ID: h.appId,
		})
	path = pathlib.Join(path, string(key))
	err = h.client.Put(path, facts)
	return
}
