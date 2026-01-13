package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// JobFunction REST resource.
type JobFunction api.JobFunction

// With updates the resource with the model.
func (r *JobFunction) With(m *model.JobFunction) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, Ref{ID: s.ID, Name: s.Name})
	}
}

// Model builds a model.
func (r *JobFunction) Model() (m *model.JobFunction) {
	m = &model.JobFunction{
		Name: r.Name,
	}
	m.ID = r.ID

	return
}
