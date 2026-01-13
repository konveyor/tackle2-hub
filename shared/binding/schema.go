package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Schema API.
type Schema struct {
	client *Client
}

// Get returns a schema by name.
func (h *Schema) Get(name string) (r *api.Schema, err error) {
	r = &api.Schema{}
	path := Path(api.SchemasGetRoute).Inject(Params{api.Name: name})
	err = h.client.Get(path, r)
	return
}

// Find returns a schema by domain, variant, subject.
func (h *Schema) Find(domain, variant, subject string) (r *api.LatestSchema, err error) {
	r = &api.LatestSchema{}
	params := Params{
		api.Domain:  domain,
		api.Variant: variant,
		api.Subject: subject,
	}
	path := Path(api.SchemaFindRoute).Inject(params)
	err = h.client.Get(path, r)
	return
}
