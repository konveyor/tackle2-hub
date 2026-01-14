package resource

import (
	"github.com/konveyor/tackle2-hub/internal/jsd"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Document REST resource.
type Document = api.Document

// Schema REST resource.
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

// Version REST resource.
type Version api.Version

func (r *Version) With(m *jsd.Version) {
	r.ID = m.ID
	r.Migration = m.Migration
	r.Definition = m.Definition
}
