package resource

import (
	"sort"

	"github.com/konveyor/tackle2-hub/internal/api/jsd"
	"github.com/konveyor/tackle2-hub/internal/assessment"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Tag Sources
const (
	SourceAssessment = "assessment"
	SourceArchetype  = "archetype"
)

// Application REST resource.
type Application api.Application

// With updates the resource using the model.
func (r *Application) With(m *model.Application, tags []AppTag, identities []IdentityRef) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Bucket = refPtr(m.BucketID, m.Bucket)
	r.Comments = m.Comments
	r.Binary = m.Binary
	if m.Coordinates != nil {
		d := jsd.Document{}
		d.With(m.Coordinates)
		r.Coordinates = &api.Document{
			Content: api.Map(d.Content),
			Schema:  d.Schema,
		}
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
		r.Review = &Ref{
			ID:   m.Review.ID,
			Name: "",
		}
	}
	r.BusinessService = refPtr(m.BusinessServiceID, m.BusinessService)
	r.Identities = identities
	r.Tags = []TagRef{}
	for i := range tags {
		r.Tags = append(r.Tags, TagRef{
			ID:      tags[i].TagID,
			Name:    tags[i].Tag.Name,
			Source:  tags[i].Source,
			Virtual: false,
		})
	}
	r.Owner = refPtr(m.OwnerID, m.Owner)
	r.Contributors = []Ref{}
	for _, c := range m.Contributors {
		r.Contributors = append(
			r.Contributors,
			Ref{
				ID:   c.ID,
				Name: c.Name,
			})
	}
	r.MigrationWave = refPtr(m.MigrationWaveID, m.MigrationWave)
	r.Platform = refPtr(m.PlatformID, m.Platform)
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		r.Assessments = append(r.Assessments, Ref{
			ID:   a.ID,
			Name: "",
		})
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
		r.Tags = append(r.Tags, TagRef{
			ID:      t.ID,
			Name:    t.Name,
			Source:  source,
			Virtual: true,
		})
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
		r.Archetypes = append(r.Archetypes, Ref{
			ID:   a.ID,
			Name: a.Name,
		})
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
		d := jsd.Document{
			Content: jsd.Map(r.Coordinates.Content),
			Schema:  r.Coordinates.Schema,
		}
		m.Coordinates = d.Model()
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

// Fact REST resource.
type Fact api.Fact

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

// FactKey REST resource.
type FactKey = api.FactKey

// TagMap REST resource.
type TagMap map[uint][]AppTag

// Set the Application.Tags.
func (r TagMap) Set(m *model.Application) {
	tags := r[m.ID]
	m.Tags = make([]model.Tag, 0, len(tags))
	for _, ref := range tags {
		m.Tags = append(m.Tags, *ref.Tag)
	}
}

// AppTag REST resource.
type AppTag struct {
	ApplicationID uint
	TagID         uint
	Source        string
	Tag           *model.Tag
}

func (r *AppTag) with(m *model.ApplicationTag) {
	r.ApplicationID = m.ApplicationID
	r.Source = m.Source
	r.Tag = &m.Tag
	r.TagID = m.TagID
}

func (r *AppTag) WithRef(m *TagRef) {
	r.Source = m.Source
	r.Tag = &model.Tag{}
	r.Tag.ID = m.ID
}

// IdentityMap REST resource.
type IdentityMap map[IdentityRef]byte

// With updates the map.
func (r IdentityMap) With(a *Application) {
	for _, ref := range a.Identities {
		r[ref] = 0
	}
}

// List returns a list of refs.
func (r IdentityMap) List() (refs []IdentityRef) {
	for ref, _ := range r {
		refs = append(refs, ref)
	}
	return
}
