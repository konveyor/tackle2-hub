package resource

import (
	"github.com/konveyor/tackle2-hub/jsd"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Document type alias to shared API.
type Document = api.Document

// Schema type alias to shared API.
type Schema api.Schema

func (r *Schema) With(m jsd.Schema) {
	r.Name = m.Name
	r.Domain = m.Domain
	r.Variant = m.Variant
	r.Subject = m.Subject
	r.Versions = []api.Version{}
	for _, v := range m.Versions {
		jv := Version{}
		jv.With(&v)
		r.Versions = append(
			r.Versions,
			api.Version(jv))
	}
}

// Version type alias to shared API.
type Version api.Version

func (r *Version) With(m *jsd.Version) {
	r.ID = m.ID
	r.Migration = m.Migration
	r.Definition = m.Definition
}
