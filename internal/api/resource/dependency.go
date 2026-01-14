package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Dependency REST resource.
type Dependency api.Dependency

// With updates the resource using the model.
func (r *Dependency) With(m *model.Dependency) {
	baseWith(&r.Resource, &m.Model)
	r.To = ref(m.ToID, m.To)
	r.From = ref(m.FromID, m.From)
}

// Model builds a model.Dependency.
func (r *Dependency) Model() (m *model.Dependency) {
	m = &model.Dependency{
		ToID:   r.To.ID,
		FromID: r.From.ID,
	}
	return
}
