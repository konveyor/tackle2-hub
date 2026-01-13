package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Review REST resource.
type Review api.Review

// With updates the resource with the model.
func (r *Review) With(m *model.Review) {
	baseWith(&r.Resource, &m.Model)
	r.BusinessCriticality = m.BusinessCriticality
	r.EffortEstimate = m.EffortEstimate
	r.ProposedAction = m.ProposedAction
	r.WorkPriority = m.WorkPriority
	r.Comments = m.Comments
	r.Application = refPtr(m.ApplicationID, m.Application)
	r.Archetype = refPtr(m.ArchetypeID, m.Archetype)
}

// Model builds a model.
func (r *Review) Model() (m *model.Review) {
	m = &model.Review{
		BusinessCriticality: r.BusinessCriticality,
		EffortEstimate:      r.EffortEstimate,
		ProposedAction:      r.ProposedAction,
		WorkPriority:        r.WorkPriority,
		Comments:            r.Comments,
	}
	m.ID = r.ID
	if r.Application != nil {
		m.ApplicationID = &r.Application.ID
	} else if r.Archetype != nil {
		m.ArchetypeID = &r.Archetype.ID
	}
	return
}

// CopyRequest REST resource.
type CopyRequest = api.CopyRequest
