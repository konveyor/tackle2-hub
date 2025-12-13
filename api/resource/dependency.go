package resource

import "github.com/konveyor/tackle2-hub/model"

// Dependency REST resource.
type Dependency struct {
	Resource `yaml:",inline"`
	To       Ref `json:"to"`
	From     Ref `json:"from"`
}

// With updates the resource using the model.
func (r *Dependency) With(m *model.Dependency) {
	r.Resource.With(&m.Model)
	r.To = r.ref(m.ToID, m.To)
	r.From = r.ref(m.FromID, m.From)
}

// Model builds a model.Dependency.
func (r *Dependency) Model() (m *model.Dependency) {
	m = &model.Dependency{
		ToID:   r.To.ID,
		FromID: r.From.ID,
	}
	return
}
