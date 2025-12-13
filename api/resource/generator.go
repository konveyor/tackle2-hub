package resource

import "github.com/konveyor/tackle2-hub/model"

// Generator REST resource.
type Generator struct {
	Resource    `yaml:",inline"`
	Kind        string      `json:"kind" binding:"required"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty" yaml:",omitempty"`
	Repository  *Repository `json:"repository"`
	Params      Map         `json:"params"`
	Values      Map         `json:"values"`
	Identity    *Ref        `json:"identity,omitempty" yaml:",omitempty"`
	Profiles    []Ref       `json:"profiles"`
}

// With updates the resource with the model.
func (r *Generator) With(m *model.Generator) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Description = m.Description
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	r.Params = m.Params
	r.Values = m.Values
	if m.Repository != (model.Repository{}) {
		repository := Repository(m.Repository)
		r.Repository = &repository
	}
	r.Profiles = make([]Ref, 0, len(m.Profiles))
	for _, p := range m.Profiles {
		r.Profiles = append(r.Profiles, r.ref(p.ID, &p))
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
