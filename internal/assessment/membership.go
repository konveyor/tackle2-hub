package assessment

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
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
	err = m.cacheArchetypeMembers(db)
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

// Archetypes returns the most specific archetypes.
// Algorithm:
// 1. Build a tag set for the application.
// 2. Identify potential archetypes whose tags are all contained in the application's tags.
// 3. Incrementally filter candidates to keep only the most specific archetypes:
//   - Discard any candidate that is a subset of an existing archetype (broader match).
//   - Replace any existing archetypes that are subsets of the candidate (narrower match).
//
// 4. Record the application as a member of each resulting archetype.
// This ensures that each application is mapped only to the narrowest matching archetypes.
func (r *MembershipResolver) Archetypes(m Application) (archetypes []Archetype, err error) {
	appTags := NewSet()
	for _, t := range m.Tags {
		appTags.Add(t.ID)
	}
	for _, a := range r.archetypes {
		aTags := r.tagSets[a.ID]
		if !appTags.Superset(aTags, false) {
			continue
		}
		dominated := false
		kept := archetypes[:0]
		for _, existing := range archetypes {
			existingTags := r.tagSets[existing.ID]
			if aTags.Subset(existingTags, true) {
				dominated = true
				break
			}
			if existingTags.Subset(aTags, true) {
				continue
			}
			kept = append(kept, existing)
		}
		archetypes = kept
		if !dominated {
			archetypes = append(archetypes, a)
		}
	}
	for _, a := range archetypes {
		r.archetypeMembers[a.ID] = append(r.archetypeMembers[a.ID], m)
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
	type M struct {
		AppId      uint
		TagId      uint
		TagName    string
		CategoryId uint
	}
	db = db.Select(
		"a.id         AppId",
		"t.id         TagId",
		"t.name       TagName",
		"t.categoryId CategoryId")
	db = db.Table("application a")
	db = db.Joins("JOIN applicationTags j ON j.applicationId = a.id")
	db = db.Joins("JOIN tag t ON t.id = j.tagId")
	db = db.Order("a.id")
	cursor, err := db.Rows()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = cursor.Close()
	}()
	application := model.Application{}
	for cursor.Next() {
		var m M
		err = db.ScanRows(cursor, &m)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if m.AppId != application.ID {
			if application.ID > 0 {
				a := Application{Application: &application}
				_, err = r.Archetypes(a)
				if err != nil {
					return
				}
			}
			application.ID = m.AppId
			application.Tags = nil
		}
		tag := model.Tag{}
		tag.ID = m.TagId
		tag.Name = m.TagName
		tag.CategoryID = m.CategoryId
		application.Tags = append(application.Tags, tag)

	}
	// Last
	if application.ID > 0 {
		a := Application{Application: &application}
		_, err = r.Archetypes(a)
		if err != nil {
			return
		}
	}

	r.membersCached = true

	return
}
