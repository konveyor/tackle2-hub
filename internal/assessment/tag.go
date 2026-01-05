package assessment

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewTagResolver builds a TagResolver.
func NewTagResolver(db *gorm.DB) (t *TagResolver, err error) {
	t = &TagResolver{
		db: db,
	}
	err = t.cacheTags()
	return
}

// TagResolver resolves CategorizedTags to Tag models.
type TagResolver struct {
	cache map[string]map[string]*model.Tag
	db    *gorm.DB
}

// Resolve a category and tag name to a Tag model.
func (r *TagResolver) Resolve(category string, tag string) (t *model.Tag, found bool) {
	t, found = r.cache[category][tag]
	return
}

// Assessment returns all the Tag models that should be applied from the assessment.
func (r *TagResolver) Assessment(assessment Assessment) (tags []model.Tag) {
	for _, t := range assessment.Tags() {
		tag, found := r.Resolve(t.Category, t.Tag)
		if found {
			tags = append(tags, *tag)
		}
	}
	return
}

func (r *TagResolver) cacheTags() (err error) {
	r.cache = make(map[string]map[string]*model.Tag)

	categories := []model.TagCategory{}
	result := r.db.Preload(clause.Associations).Find(&categories)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}

	for _, c := range categories {
		r.cache[c.Name] = make(map[string]*model.Tag)
		for i := range c.Tags {
			t := &c.Tags[i]
			r.cache[c.Name][t.Name] = t
		}
	}

	return
}
