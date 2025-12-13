package resource

import "github.com/konveyor/tackle2-hub/model"

// Manifest REST resource.
type Manifest struct {
	Resource    `yaml:",inline"`
	Content     Map `json:"content"`
	Secret      Map `json:"secret,omitempty" yaml:"secret,omitempty"`
	Application Ref `json:"application"`
}

// With updates the resource with the model.
func (r *Manifest) With(m *model.Manifest) {
	r.Resource.With(&m.Model)
	r.Content = m.Content
	r.Secret = m.Secret
	ref := Ref{}
	ref.With(m.ApplicationID, "")
	r.Application = ref
}

// Model builds a model.
func (r *Manifest) Model() (m *model.Manifest) {
	m = &model.Manifest{}
	m.ID = r.ID
	m.Content = r.Content
	m.Secret = r.Secret
	m.ApplicationID = r.Application.ID
	return
}
