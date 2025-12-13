package resource

import "github.com/konveyor/tackle2-hub/model"

// StakeholderGroup REST resource.
type StakeholderGroup struct {
	Resource       `yaml:",inline"`
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Stakeholders   []Ref  `json:"stakeholders"`
	MigrationWaves []Ref  `json:"migrationWaves" yaml:"migrationWaves"`
}

// With updates the resource with the model.
func (r *StakeholderGroup) With(m *model.StakeholderGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
	r.MigrationWaves = []Ref{}
	for _, w := range m.MigrationWaves {
		ref := Ref{}
		ref.With(w.ID, w.Name)
		r.MigrationWaves = append(r.MigrationWaves, ref)
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
