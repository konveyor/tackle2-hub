package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Schema API.
type Schema struct {
	client *Client
}

// Get returns a schema by name.
func (h *Schema) Get(name string) (r *api2.Schema, err error) {
	r = &api2.Schema{}
	path := Path(api2.SchemasGetRoute).Inject(Params{api2.Name: name})
	err = h.client.Get(path, r)
	return
}

// Find returns a schema by domain, variant, subject.
func (h *Schema) Find(domain, variant, subject string) (r *api2.LatestSchema, err error) {
	r = &api2.LatestSchema{}
	params := Params{
		api2.Domain:  domain,
		api2.Variant: variant,
		api2.Subject: subject,
	}
	path := Path(api2.SchemaFindRoute).Inject(params)
	err = h.client.Get(path, r)
	return
}
