package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Generator REST resource.
type Generator api.Generator

// With updates the resource with the model.
func (r *Generator) With(m *model.Generator) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Description = m.Description
	r.Identity = refPtr(m.IdentityID, m.Identity)
	r.Params = m.Params
	r.Values = m.Values
	if m.Repository != (model.Repository{}) {
		repository := Repository(m.Repository)
		r.Repository = &repository
	}
	r.Profiles = make([]Ref, 0, len(m.Profiles))
	for _, p := range m.Profiles {
		r.Profiles = append(r.Profiles, ref(p.ID, &p))
	}
}

// Model builds a model.
func (r *Generator) Model() (m *model.Generator) {
	m = &model.Generator{}
	m.ID = r.ID
	m.Kind = r.Kind
	m.Name = r.Name
	m.Description = r.Description
	m.Params = r.Params
	m.Values = r.Values
	if r.Repository != nil {
		m.Repository = model.Repository(*r.Repository)
	}
	if r.Identity != nil {
		m.IdentityID = &r.Identity.ID
	}

	return
}
