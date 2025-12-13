package resource

import (
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
)

// TargetProfile REST resource.
type TargetProfile struct {
	Resource        `yaml:",inline"`
	Name            string `json:"name" binding:"required"`
	Generators      []Ref  `json:"generators"`
	AnalysisProfile *Ref   `json:"analysisProfile,omitempty" yaml:"analysisProfile,omitempty"`
}

// With updates the resource with the model.
func (r *TargetProfile) With(m *model.TargetProfile) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Generators = []Ref{}
	for _, g := range m.Generators {
		ref := Ref{}
		ref.With(g.Generator.ID, g.Generator.Name)
		r.Generators = append(r.Generators, ref)
	}
	r.AnalysisProfile = r.refPtr(
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
type Archetype struct {
	Resource          `yaml:",inline"`
	Name              string          `json:"name" yaml:"name"`
	Description       string          `json:"description" yaml:"description"`
	Comments          string          `json:"comments" yaml:"comments"`
	Tags              []TagRef        `json:"tags" yaml:"tags"`
	Criteria          []TagRef        `json:"criteria" yaml:"criteria"`
	Stakeholders      []Ref           `json:"stakeholders" yaml:"stakeholders"`
	StakeholderGroups []Ref           `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	Applications      []Ref           `json:"applications" yaml:"applications"`
	Assessments       []Ref           `json:"assessments" yaml:"assessments"`
	Assessed          bool            `json:"assessed"`
	Risk              string          `json:"risk"`
	Confidence        int             `json:"confidence"`
	Review            *Ref            `json:"review"`
	Profiles          []TargetProfile `json:"profiles" yaml:",omitempty"`
}

// With updates the resource with the model.
func (r *Archetype) With(m *model.Archetype) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Comments = m.Comments
	r.Tags = []TagRef{}
	for _, t := range m.Tags {
		ref := TagRef{}
		ref.With(t.ID, t.Name, "", false)
		r.Tags = append(r.Tags, ref)
	}
	r.Criteria = []TagRef{}
	for _, t := range m.CriteriaTags {
		ref := TagRef{}
		ref.With(t.ID, t.Name, "", false)
		r.Criteria = append(r.Criteria, ref)
	}
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, r.ref(s.ID, &s))
	}
	r.StakeholderGroups = []Ref{}
	for _, g := range m.StakeholderGroups {
		r.StakeholderGroups = append(r.StakeholderGroups, r.ref(g.ID, &g))
	}
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		r.Assessments = append(r.Assessments, r.ref(a.ID, &a))
	}
	if m.Review != nil {
		ref := &Ref{}
		ref.With(m.Review.ID, "")
		r.Review = ref
	}
	r.Risk = assessment.RiskUnassessed
	r.Profiles = []TargetProfile{}
	for _, p := range m.Profiles {
		pr := TargetProfile{}
		pr.With(&p)
		r.Profiles = append(r.Profiles, pr)
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
		ref := Ref{}
		ref.With(apps[i].ID, apps[i].Name)
		r.Applications = append(r.Applications, ref)
	}
	for _, t := range resolver.AssessmentTags() {
		ref := TagRef{}
		ref.With(t.ID, t.Name, SourceAssessment, true)
		r.Tags = append(r.Tags, ref)
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
		pm := p.Model()
		m.Profiles = append(
			m.Profiles,
			*pm)
	}

	return
}
