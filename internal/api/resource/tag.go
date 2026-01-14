package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Tag REST resource.
type Tag api.Tag

// With updates the resource with the model.
func (r *Tag) With(m *model.Tag) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Category = ref(m.CategoryID, &m.Category)
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
