package resource

import (
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

type Questionnaire api.Questionnaire

// With updates the resource with the model.
func (r *Questionnaire) With(m *model.Questionnaire) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Required = m.Required
	r.Builtin = m.Builtin()
	r.Sections = []api.Section{}
	for _, s := range m.Sections {
		sect := Section{}
		sect.With(&s)
		r.Sections = append(r.Sections, api.Section(sect))
	}
	r.Thresholds = api.Thresholds(m.Thresholds)
	r.RiskMessages = api.RiskMessages(m.RiskMessages)
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
		sect := Section(s)
		m.Sections = append(m.Sections, *sect.Model())
	}
	m.Thresholds = model.Thresholds(r.Thresholds)
	m.RiskMessages = model.RiskMessages(r.RiskMessages)

	return
}
