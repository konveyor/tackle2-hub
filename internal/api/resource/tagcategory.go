package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// TagCategory REST resource.
type TagCategory api.TagCategory

// With updates the resource with the model.
func (r *TagCategory) With(m *model.TagCategory) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Color = m.Color
	r.Tags = []Ref{}
	for _, tag := range m.Tags {
		r.Tags = append(r.Tags, Ref{ID: tag.ID, Name: tag.Name})
	}
}

// Model builds a model.
func (r *TagCategory) Model() (m *model.TagCategory) {
	m = &model.TagCategory{
		Name:  r.Name,
		Color: r.Color,
	}
	m.ID = r.ID
	return
}
