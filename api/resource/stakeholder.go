package resource

import "github.com/konveyor/tackle2-hub/model"

// Stakeholder REST resource.
type Stakeholder struct {
	Resource         `yaml:",inline"`
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	JobFunction      *Ref   `json:"jobFunction" yaml:"jobFunction"`
	Groups           []Ref  `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	BusinessServices []Ref  `json:"businessServices" yaml:"businessServices"`
	Owns             []Ref  `json:"owns"`
	Contributes      []Ref  `json:"contributes"`
	MigrationWaves   []Ref  `json:"migrationWaves" yaml:"migrationWaves"`
}

// With updates the resource with the model.
func (r *Stakeholder) With(m *model.Stakeholder) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Email = m.Email
	r.JobFunction = r.refPtr(m.JobFunctionID, m.JobFunction)
	r.Groups = []Ref{}
	for _, g := range m.Groups {
		ref := Ref{}
		ref.With(g.ID, g.Name)
		r.Groups = append(r.Groups, ref)
	}
	r.BusinessServices = []Ref{}
	for _, bs := range m.BusinessServices {
		ref := Ref{}
		ref.With(bs.ID, bs.Name)
		r.BusinessServices = append(r.BusinessServices, ref)
	}
	r.Owns = []Ref{}
	for _, a := range m.Owns {
		ref := Ref{}
		ref.With(a.ID, a.Name)
		r.Owns = append(r.Owns, ref)
	}
	r.Contributes = []Ref{}
	for _, a := range m.Contributes {
		ref := Ref{}
		ref.With(a.ID, a.Name)
		r.Contributes = append(r.Contributes, ref)
	}
	r.MigrationWaves = []Ref{}
	for _, mw := range m.MigrationWaves {
		ref := Ref{}
		ref.With(mw.ID, mw.Name)
		r.MigrationWaves = append(r.MigrationWaves, ref)
	}
}

// Model builds a model.
func (r *Stakeholder) Model() (m *model.Stakeholder) {
	m = &model.Stakeholder{
		Name:  r.Name,
		Email: r.Email,
	}
	m.ID = r.ID
	if r.JobFunction != nil {
		m.JobFunctionID = &r.JobFunction.ID
	}
	for _, ref := range r.Groups {
		m.Groups = append(
			m.Groups,
			model.StakeholderGroup{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.BusinessServices {
		m.BusinessServices = append(
			m.BusinessServices,
			model.BusinessService{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.Owns {
		m.Owns = append(
			m.Owns,
			model.Application{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.Contributes {
		m.Contributes = append(
			m.Contributes,
			model.Application{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.MigrationWaves {
		m.MigrationWaves = append(
			m.MigrationWaves,
			model.MigrationWave{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}
