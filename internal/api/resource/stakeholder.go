package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Stakeholder REST resource.
type Stakeholder api.Stakeholder

// With updates the resource with the model.
func (r *Stakeholder) With(m *model.Stakeholder) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Email = m.Email
	r.JobFunction = refPtr(m.JobFunctionID, m.JobFunction)
	r.Groups = []Ref{}
	for _, g := range m.Groups {
		r.Groups = append(r.Groups, Ref{ID: g.ID, Name: g.Name})
	}
	r.BusinessServices = []Ref{}
	for _, bs := range m.BusinessServices {
		r.BusinessServices = append(r.BusinessServices, Ref{ID: bs.ID, Name: bs.Name})
	}
	r.Owns = []Ref{}
	for _, a := range m.Owns {
		r.Owns = append(r.Owns, Ref{ID: a.ID, Name: a.Name})
	}
	r.Contributes = []Ref{}
	for _, a := range m.Contributes {
		r.Contributes = append(r.Contributes, Ref{ID: a.ID, Name: a.Name})
	}
	r.MigrationWaves = []Ref{}
	for _, mw := range m.MigrationWaves {
		r.MigrationWaves = append(r.MigrationWaves, Ref{ID: mw.ID, Name: mw.Name})
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
