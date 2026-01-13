package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// StakeholderGroup REST resource.
type StakeholderGroup api.StakeholderGroup

// With updates the resource with the model.
func (r *StakeholderGroup) With(m *model.StakeholderGroup) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, Ref{ID: s.ID, Name: s.Name})
	}
	r.MigrationWaves = []Ref{}
	for _, w := range m.MigrationWaves {
		r.MigrationWaves = append(r.MigrationWaves, Ref{ID: w.ID, Name: w.Name})
	}
}

// Model builds a model.
func (r *StakeholderGroup) Model() (m *model.StakeholderGroup) {
	m = &model.StakeholderGroup{
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	for _, s := range r.Stakeholders {
		m.Stakeholders = append(m.Stakeholders, model.Stakeholder{Model: model.Model{ID: s.ID}})
	}
	for _, w := range r.MigrationWaves {
		m.MigrationWaves = append(m.MigrationWaves, model.MigrationWave{Model: model.Model{ID: w.ID}})
	}
	return
}
