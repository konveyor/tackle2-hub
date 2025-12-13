package resource

import "github.com/konveyor/tackle2-hub/model"

// Tag REST resource.
type Tag struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Category Ref    `json:"category" binding:"required"`
}

// With updates the resource with the model.
func (r *Tag) With(m *model.Tag) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Category = r.ref(m.CategoryID, &m.Category)
}

// Model builds a model.
func (r *Tag) Model() (m *model.Tag) {
	m = &model.Tag{
		Name:       r.Name,
		CategoryID: r.Category.ID,
	}
	m.ID = r.ID
	return
}
