package assessment

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewMembershipResolver builds a MembershipResolver.
func NewMembershipResolver(db *gorm.DB) (m *MembershipResolver, err error) {
	m = &MembershipResolver{}
	m.tagSets = make(map[uint]Set)
	m.archetypeMembers = make(map[uint][]Application)
	err = m.cacheArchetypes(db)
	if err != nil {
		return
	}
	return
}

// MembershipResolver resolves archetype membership.
type MembershipResolver struct {
	archetypes       []Archetype
	tagSets          map[uint]Set
	archetypeMembers map[uint][]Application
	membersCached    bool
	archetypesCached bool
}

// Applications returns the list of applications that are members of the given archetype.
func (r *MembershipResolver) Applications(m Archetype) (applications []Application, err error) {
	applications = r.archetypeMembers[m.ID]
	return
}

// Archetypes returns the list of archetypes that the application is a member of.
func (r *MembershipResolver) Archetypes(m Application) (archetypes []Archetype, err error) {
	appTags := NewSet()
	for _, t := range m.Tags {
		appTags.Add(t.ID)
	}

	matches := []Archetype{}
	for _, a := range r.archetypes {
		if appTags.Superset(r.tagSets[a.ID], false) {
			matches = append(matches, a)
		}
	}

	// throw away any archetypes that are a subset of
	// another archetype-- only keep the most specific matches.
loop:
	for _, a1 := range matches {
		for _, a2 := range matches {
			if a1.ID == a2.ID {
				continue
			}
			a1tags := r.tagSets[a1.ID]
			a2tags := r.tagSets[a2.ID]
			if a1tags.Subset(a2tags, true) {
				continue loop
			}
		}
		archetypes = append(archetypes, a1)
		r.archetypeMembers[a1.ID] = append(r.archetypeMembers[a1.ID], m)
	}

	return
}

func (r *MembershipResolver) cacheArchetypes(db *gorm.DB) (err error) {
	if r.archetypesCached {
		return
	}

	list := []model.Archetype{}
	db = db.Preload(clause.Associations)
	db = db.Preload("Assessments.Stakeholders")
	db = db.Preload("Assessments.StakeholderGroups")
	result := db.Find(&list)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}

	for i := range list {
		a := Archetype{}
		a.With(&list[i])
		r.archetypes = append(r.archetypes, a)
		set := NewSet()
		for _, t := range a.CriteriaTags {
			set.Add(t.ID)
		}
		r.tagSets[a.ID] = set
	}
	r.archetypesCached = true

	return
}

func (r *MembershipResolver) cacheArchetypeMembers(db *gorm.DB) (err error) {
	if r.membersCached {
		return
	}
	list := []model.Application{}
	result := db.Preload("Tags").Find(&list)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}

	for i := range list {
		a := Application{}
		a.With(&list[i])
		_, aErr := r.Archetypes(a)
		if aErr != nil {
			err = aErr
			return
		}
	}
	r.membersCached = true

	return
}
