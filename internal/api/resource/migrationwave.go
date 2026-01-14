package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// MigrationWave REST resource.
type MigrationWave api.MigrationWave

// With updates the resource using the model.
func (r *MigrationWave) With(m *model.MigrationWave) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.StartDate = m.StartDate
	r.EndDate = m.EndDate
	r.Applications = []Ref{}
	for _, app := range m.Applications {
		r.Applications = append(r.Applications, Ref{ID: app.ID, Name: app.Name})
	}
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, Ref{ID: s.ID, Name: s.Name})
	}
	r.StakeholderGroups = []Ref{}
	for _, sg := range m.StakeholderGroups {
		r.StakeholderGroups = append(r.StakeholderGroups, Ref{ID: sg.ID, Name: sg.Name})
	}
}

// Model builds a model.
func (r *MigrationWave) Model() (m *model.MigrationWave) {
	m = &model.MigrationWave{
		Name:      r.Name,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
	}
	m.ID = r.ID
	for _, ref := range r.Applications {
		m.Applications = append(
			m.Applications,
			model.Application{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.Stakeholders {
		m.Stakeholders = append(
			m.Stakeholders,
			model.Stakeholder{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.StakeholderGroups {
		m.StakeholderGroups = append(
			m.StakeholderGroups,
			model.StakeholderGroup{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}
