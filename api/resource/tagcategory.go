package resource

import "github.com/konveyor/tackle2-hub/model"

// TagCategory REST resource.
type TagCategory struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Color    string `json:"colour,omitempty" yaml:"colour,omitempty"`
	Rank     *uint  `json:"rank,omitempty" yaml:"rank,omitempty"`
	Tags     []Ref  `json:"tags"`
}

// With updates the resource with the model.
func (r *TagCategory) With(m *model.TagCategory) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Color = m.Color
	r.Tags = []Ref{}
	for _, tag := range m.Tags {
		ref := Ref{}
		ref.With(tag.ID, tag.Name)
		r.Tags = append(r.Tags, ref)
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
