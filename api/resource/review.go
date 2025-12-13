package resource

import "github.com/konveyor/tackle2-hub/model"

// Review REST resource.
type Review struct {
	Resource            `yaml:",inline"`
	BusinessCriticality uint   `json:"businessCriticality" yaml:"businessCriticality"`
	EffortEstimate      string `json:"effortEstimate" yaml:"effortEstimate"`
	ProposedAction      string `json:"proposedAction" yaml:"proposedAction"`
	WorkPriority        uint   `json:"workPriority" yaml:"workPriority"`
	Comments            string `json:"comments"`
	Application         *Ref   `json:"application,omitempty" binding:"required_without=Archetype,excluded_with=Archetype"`
	Archetype           *Ref   `json:"archetype,omitempty" binding:"required_without=Application,excluded_with=Application"`
}

// With updates the resource with the model.
func (r *Review) With(m *model.Review) {
	r.Resource.With(&m.Model)
	r.BusinessCriticality = m.BusinessCriticality
	r.EffortEstimate = m.EffortEstimate
	r.ProposedAction = m.ProposedAction
	r.WorkPriority = m.WorkPriority
	r.Comments = m.Comments
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Archetype = r.refPtr(m.ArchetypeID, m.Archetype)
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
