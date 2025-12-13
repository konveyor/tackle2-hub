package resource

import "github.com/konveyor/tackle2-hub/model"

type Questionnaire struct {
	Resource     `yaml:",inline"`
	Name         string       `json:"name" yaml:"name" binding:"required"`
	Description  string       `json:"description" yaml:"description"`
	Required     bool         `json:"required" yaml:"required"`
	Sections     []Section    `json:"sections" yaml:"sections" binding:"required,min=1,dive"`
	Thresholds   Thresholds   `json:"thresholds" yaml:"thresholds" binding:"required"`
	RiskMessages RiskMessages `json:"riskMessages" yaml:"riskMessages" binding:"required"`
	Builtin      bool         `json:"builtin,omitempty" yaml:"builtin,omitempty"`
}

// With updates the resource with the model.
func (r *Questionnaire) With(m *model.Questionnaire) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Required = m.Required
	r.Builtin = m.Builtin()
	r.Sections = []Section{}
	for _, s := range m.Sections {
		r.Sections = append(r.Sections, Section(s))
	}
	r.Thresholds = Thresholds(m.Thresholds)
	r.RiskMessages = RiskMessages(m.RiskMessages)
}

// Model builds a model.
func (r *Questionnaire) Model() (m *model.Questionnaire) {
	m = &model.Questionnaire{
		Name:        r.Name,
		Description: r.Description,
		Required:    r.Required,
	}
	m.ID = r.ID
	for _, s := range r.Sections {
		m.Sections = append(m.Sections, model.Section(s))
	}
	m.Thresholds = model.Thresholds(r.Thresholds)
	m.RiskMessages = model.RiskMessages(r.RiskMessages)

	return
}
