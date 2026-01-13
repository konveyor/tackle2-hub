package resource

import (
	"github.com/konveyor/tackle2-hub/internal/assessment"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// TargetProfile REST resource.
type TargetProfile api.TargetProfile

// With updates the resource with the model.
func (r *TargetProfile) With(m *model.TargetProfile) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Generators = []Ref{}
	for _, g := range m.Generators {
		r.Generators = append(r.Generators, Ref{ID: g.Generator.ID, Name: g.Generator.Name})
	}
	r.AnalysisProfile = refPtr(
		m.AnalysisProfileID,
		m.AnalysisProfile)
}

// Model builds a model from the resource.
func (r *TargetProfile) Model() (m *model.TargetProfile) {
	m = &model.TargetProfile{}
	m.ID = r.ID
	m.Name = r.Name
	for _, ref := range r.Generators {
		g := model.ProfileGenerator{}
		g.GeneratorID = ref.ID
		g.TargetProfileID = m.ID
		m.Generators = append(
			m.Generators,
			g)
	}
	if r.AnalysisProfile != nil {
		m.AnalysisProfileID = &r.AnalysisProfile.ID
	}
	return
}

// Archetype REST resource.
type Archetype api.Archetype

// With updates the resource with the model.
func (r *Archetype) With(m *model.Archetype) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Comments = m.Comments
	r.Tags = []TagRef{}
	for _, t := range m.Tags {
		r.Tags = append(r.Tags, TagRef{ID: t.ID, Name: t.Name, Source: "", Virtual: false})
	}
	r.Criteria = []TagRef{}
	for _, t := range m.CriteriaTags {
		r.Criteria = append(r.Criteria, TagRef{ID: t.ID, Name: t.Name, Source: "", Virtual: false})
	}
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, ref(s.ID, &s))
	}
	r.StakeholderGroups = []Ref{}
	for _, g := range m.StakeholderGroups {
		r.StakeholderGroups = append(r.StakeholderGroups, ref(g.ID, &g))
	}
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		r.Assessments = append(r.Assessments, ref(a.ID, &a))
	}
	if m.Review != nil {
		r.Review = &Ref{ID: m.Review.ID, Name: ""}
	}
	r.Risk = assessment.RiskUnassessed
	r.Profiles = []api.TargetProfile{}
	for _, p := range m.Profiles {
		pr := TargetProfile{}
		pr.With(&p)
		r.Profiles = append(r.Profiles, api.TargetProfile(pr))
	}
}

// WithResolver uses an ArchetypeResolver to update the resource with
// values derived from the archetype's assessments.
func (r *Archetype) WithResolver(resolver *assessment.ArchetypeResolver) (err error) {
	r.Assessed = resolver.Assessed()
	if r.Assessed {
		r.Risk = resolver.Risk()
		r.Confidence = resolver.Confidence()
	}
	apps, err := resolver.Applications()
	for i := range apps {
		r.Applications = append(r.Applications, Ref{ID: apps[i].ID, Name: apps[i].Name})
	}
	for _, t := range resolver.AssessmentTags() {
		r.Tags = append(r.Tags, TagRef{ID: t.ID, Name: t.Name, Source: SourceAssessment, Virtual: true})
	}
	return
}

// Model builds a model from the resource.
func (r *Archetype) Model() (m *model.Archetype) {
	m = &model.Archetype{
		Name:        r.Name,
		Description: r.Description,
		Comments:    r.Comments,
	}
	m.ID = r.ID
	for _, ref := range r.Tags {
		if ref.Virtual {
			continue
		}
		m.Tags = append(
			m.Tags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.Criteria {
		m.CriteriaTags = append(
			m.CriteriaTags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.Stakeholders {
		m.Stakeholders = append(
			m.Stakeholders,
			model.Stakeholder{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.StakeholderGroups {
		m.StakeholderGroups = append(
			m.StakeholderGroups,
			model.StakeholderGroup{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, p := range r.Profiles {
		tp := TargetProfile(p)
		pm := tp.Model()
		m.Profiles = append(
			m.Profiles,
			*pm)
	}

	return
}
