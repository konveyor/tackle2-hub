package resource

import (
	"sort"

	"github.com/konveyor/tackle2-hub/api/jsd"
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
)

// Application REST resource.
type Application struct {
	Resource        `yaml:",inline"`
	Name            string        `json:"name" binding:"required"`
	Description     string        `json:"description"`
	Bucket          *Ref          `json:"bucket"`
	Repository      *Repository   `json:"repository"`
	Assets          *Repository   `json:"assets"`
	Binary          string        `json:"binary"`
	Coordinates     *jsd.Document `json:"coordinates"`
	Review          *Ref          `json:"review"`
	Comments        string        `json:"comments"`
	Identities      []IdentityRef `json:"identities"`
	Tags            []TagRef      `json:"tags"`
	BusinessService *Ref          `json:"businessService" yaml:"businessService"`
	Owner           *Ref          `json:"owner"`
	Contributors    []Ref         `json:"contributors"`
	MigrationWave   *Ref          `json:"migrationWave" yaml:"migrationWave"`
	Platform        *Ref          `json:"platform"`
	Archetypes      []Ref         `json:"archetypes"`
	Assessments     []Ref         `json:"assessments"`
	Manifests       []Ref         `json:"manifests"`
	Assessed        bool          `json:"assessed"`
	Risk            string        `json:"risk"`
	Confidence      int           `json:"confidence"`
	Effort          int           `json:"effort"`
}

// With updates the resource using the model.
func (r *Application) With(m *model.Application, tags []AppTag, identities []IdentityRef) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.Comments = m.Comments
	r.Binary = m.Binary
	if m.Coordinates != nil {
		d := jsd.Document{}
		d.With(m.Coordinates)
		r.Coordinates = &d
	}
	if m.Repository != (model.Repository{}) {
		repo := Repository(m.Repository)
		r.Repository = &repo
	}
	if m.Assets != (model.Repository{}) {
		repo := Repository(m.Assets)
		r.Assets = &repo
	}
	if m.Review != nil {
		ref := &Ref{}
		ref.With(m.Review.ID, "")
		r.Review = ref
	}
	r.BusinessService = r.refPtr(m.BusinessServiceID, m.BusinessService)
	r.Identities = identities
	r.Tags = []TagRef{}
	for i := range tags {
		ref := TagRef{}
		ref.With(tags[i].TagID, tags[i].Tag.Name, tags[i].Source, false)
		r.Tags = append(r.Tags, ref)
	}
	r.Owner = r.refPtr(m.OwnerID, m.Owner)
	r.Contributors = []Ref{}
	for _, c := range m.Contributors {
		ref := Ref{}
		ref.With(c.ID, c.Name)
		r.Contributors = append(
			r.Contributors,
			ref)
	}
	r.MigrationWave = r.refPtr(m.MigrationWaveID, m.MigrationWave)
	r.Platform = r.refPtr(m.PlatformID, m.Platform)
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		ref := Ref{}
		ref.With(a.ID, "")
		r.Assessments = append(r.Assessments, ref)
	}
	if len(m.Analyses) > 0 {
		sort.Slice(m.Analyses, func(i, j int) bool {
			return m.Analyses[i].ID < m.Analyses[j].ID
		})
		r.Effort = m.Analyses[len(m.Analyses)-1].Effort
	}
	r.Manifests = []Ref{}
	for _, mf := range m.Manifest {
		r.Manifests = append(r.Manifests, Ref{ID: mf.ID})
	}
	r.Risk = assessment.RiskUnassessed
}

// WithVirtualTags updates the resource with tags derived from assessments.
func (r *Application) WithVirtualTags(tags []model.Tag, source string) {
	for _, t := range tags {
		ref := TagRef{}
		ref.With(t.ID, t.Name, source, true)
		r.Tags = append(r.Tags, ref)
	}
}

// WithResolver uses an ApplicationResolver to update the resource with
// values derived from the application's assessments and archetypes.
func (r *Application) WithResolver(m *model.Application, resolver *assessment.ApplicationResolver) (err error) {
	archetypes, err := resolver.Archetypes(m)
	if err != nil {
		return
	}
	for _, a := range archetypes {
		ref := Ref{}
		ref.With(a.ID, a.Name)
		r.Archetypes = append(r.Archetypes, ref)
	}
	archetypeTags, err := resolver.ArchetypeTags(m)
	if err != nil {
		return
	}
	r.WithVirtualTags(archetypeTags, SourceArchetype)
	r.WithVirtualTags(resolver.AssessmentTags(m), SourceAssessment)
	r.Assessed, err = resolver.Assessed(m)
	if err != nil {
		return
	}
	if r.Assessed {
		r.Confidence, err = resolver.Confidence(m)
		if err != nil {
			return
		}
		r.Risk, err = resolver.Risk(m)
		if err != nil {
			return
		}
	}
	return
}

// Model builds a model.
func (r *Application) Model() (m *model.Application) {
	m = &model.Application{
		Name:        r.Name,
		Description: r.Description,
		Comments:    r.Comments,
		Binary:      r.Binary,
	}
	m.ID = r.ID
	if r.Coordinates != nil {
		d := r.Coordinates.Model()
		m.Coordinates = d
	}
	if r.Repository != nil {
		m.Repository = model.Repository(*r.Repository)
	}
	if r.Assets != nil {
		m.Assets = model.Repository(*r.Assets)
	}
	if r.BusinessService != nil {
		m.BusinessServiceID = &r.BusinessService.ID
	}
	for _, ref := range r.Identities {
		m.Identities = append(
			m.Identities,
			model.Identity{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.Tags {
		m.Tags = append(
			m.Tags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	if r.Owner != nil {
		m.OwnerID = &r.Owner.ID
	}
	for _, ref := range r.Contributors {
		m.Contributors = append(
			m.Contributors,
			model.Stakeholder{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	if r.MigrationWave != nil {
		m.MigrationWaveID = &r.MigrationWave.ID
	}
	if r.Platform != nil {
		m.PlatformID = &r.Platform.ID
	}

	return
}

// Fact REST nested resource.
type Fact struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

func (r *Fact) With(m *model.Fact) {
	r.Key = m.Key
	r.Source = m.Source
	r.Value = m.Value
}

func (r *Fact) Model() (m *model.Fact) {
	m = &model.Fact{}
	m.Key = r.Key
	m.Source = r.Source
	m.Value = r.Value
	return
}
