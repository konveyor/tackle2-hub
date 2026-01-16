package jsd

import (
	"github.com/konveyor/tackle2-hub/internal/jsd"
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	"github.com/konveyor/tackle2-hub/internal/model"
)

// Map unstructured object.
type Map json.Map

// As convert the content into the object.
// The object must be a pointer.
func (m *Map) As(object any) (err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, object)
	return
}

// Document REST nested resource.
type Document struct {
	Content Map    `json:"content" binding:"required"`
	Schema  string `json:"schema,omitempty"`
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

func (d *Document) With(md *model.Document) {
	d.Content = md.Content
	d.Schema = md.Schema
}

func (d *Document) Model() (m *model.Document) {
	m = &model.Document{}
	m.Content = d.Content
	m.Schema = d.Schema
	return
}
