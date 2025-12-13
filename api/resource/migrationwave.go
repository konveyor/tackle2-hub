package resource

import (
	"time"

	"github.com/konveyor/tackle2-hub/model"
)

// MigrationWave REST resource.
type MigrationWave struct {
	Resource          `yaml:",inline"`
	Name              string    `json:"name"`
	StartDate         time.Time `json:"startDate" yaml:"startDate" binding:"required"`
	EndDate           time.Time `json:"endDate" yaml:"endDate" binding:"required,gtfield=StartDate"`
	Applications      []Ref     `json:"applications"`
	Stakeholders      []Ref     `json:"stakeholders"`
	StakeholderGroups []Ref     `json:"stakeholderGroups" yaml:"stakeholderGroups"`
}

// With updates the resource using the model.
func (r *MigrationWave) With(m *model.MigrationWave) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.StartDate = m.StartDate
	r.EndDate = m.EndDate
	r.Applications = []Ref{}
	for _, app := range m.Applications {
		ref := Ref{}
		ref.With(app.ID, app.Name)
		r.Applications = append(r.Applications, ref)
	}
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
	r.StakeholderGroups = []Ref{}
	for _, sg := range m.StakeholderGroups {
		ref := Ref{}
		ref.With(sg.ID, sg.Name)
		r.StakeholderGroups = append(r.StakeholderGroups, ref)
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
