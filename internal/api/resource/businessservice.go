package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// BusinessService REST resource.
type BusinessService api.BusinessService

// With updates the resource with the model.
func (r *BusinessService) With(m *model.BusinessService) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Stakeholder = refPtr(m.StakeholderID, m.Stakeholder)
}

// Model builds a model.
func (r *BusinessService) Model() (m *model.BusinessService) {
	m = &model.BusinessService{
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	if r.Stakeholder != nil {
		m.StakeholderID = &r.Stakeholder.ID
	}
	return
}
