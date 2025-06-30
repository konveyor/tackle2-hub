package jsd

import (
	"github.com/konveyor/tackle2-hub/jsd"
	"github.com/konveyor/tackle2-hub/migration/json"
)

// Document REST nested resource.
type Document struct {
	Content json.Map `json:"content" binding:"required"`
	Schema  string   `json:"schema,omitempty"`
}

// Validate based on schema.
func (d *Document) Validate(m *jsd.Manager) (err error) {
	if d.Schema == "" {
		return
	}
	schema, err := m.Get(d.Schema)
	if err != nil {
		return
	}
	err = schema.Validate(d.Content)
	return
}
