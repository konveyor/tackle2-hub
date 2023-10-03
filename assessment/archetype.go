package assessment

import "github.com/konveyor/tackle2-hub/model"

//
// NewArchetypeResolver creates a new ArchetypeResolver.
func NewArchetypeResolver(archetype *model.Archetype, tags *TagResolver) (a *ArchetypeResolver) {
	a = &ArchetypeResolver{
		archetype:   archetype,
		tagResolver: tags,
	}
	return
}

//
// ArchetypeResolver wraps an Archetype model
// with assessment-related functionality.
type ArchetypeResolver struct {
	archetype   *model.Archetype
	tagResolver *TagResolver
}

//
// AssessmentTags returns the list of tags that the archetype should
// inherit from the answers given to its assessments.
func (r *ArchetypeResolver) AssessmentTags() (tags []model.Tag) {
	seenTags := make(map[uint]bool)
	for _, assessment := range r.archetype.Assessments {
		aTags := r.tagResolver.Assessment(&assessment)
		for _, t := range aTags {
			if _, found := seenTags[t.ID]; !found {
				seenTags[t.ID] = true
				tags = append(tags, t)
			}
		}
	}
	return
}
