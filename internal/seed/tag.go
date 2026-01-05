package seed

import (
	"errors"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

// TagCategory applies TagCategory seeds.
type TagCategory struct {
	categories []libseed.TagCategory
}

// With collects all the TagCategory seeds.
func (r *TagCategory) With(seed libseed.Seed) (err error) {
	items, err := seed.DecodeItems()
	if err != nil {
		return
	}
	for _, item := range items {
		r.categories = append(r.categories, item.(libseed.TagCategory))
	}
	return
}

// Apply seeds the database with TagCategories and Tags.
func (r *TagCategory) Apply(db *gorm.DB) (err error) {
	log.Info("Applying TagCategories", "count", len(r.categories))
	for i := range r.categories {
		tc := r.categories[i]
		category, found, fErr := r.find(db, "uuid = ?", tc.UUID)
		if fErr != nil {
			err = fErr
			return
		}
		// model exists and is being renamed
		if found && category.Name != tc.Name {
			// ensure that the target name is clear
			collision, collides, fErr := r.find(db, "name = ? and id != ?", tc.Name, category.ID)
			if fErr != nil {
				err = fErr
				return
			}
			if collides {
				err = r.rename(db, collision)
				if err != nil {
					return
				}
			}
		} else if !found {
			category, found, fErr = r.find(db, "name = ?", tc.Name)
			if fErr != nil {
				err = fErr
				return
			}
			// model already exists but wasn't created by the hub
			if found && category.CreateUser != "" {
				err = r.rename(db, category)
				if err != nil {
					return
				}
				found = false
			}
			if !found {
				category = &model.TagCategory{}
			}
		}

		category.Name = tc.Name
		category.UUID = &tc.UUID
		category.Color = tc.Color
		result := db.Save(&category)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}

		err = r.applyTags(db, category, tc)
		if err != nil {
			return
		}
	}
	return
}

// Seed a TagCategory's tags.
func (r *TagCategory) applyTags(db *gorm.DB, category *model.TagCategory, tc libseed.TagCategory) (err error) {
	for i := range tc.Tags {
		t := tc.Tags[i]
		tag := model.Tag{}
		result := db.First(&tag, model.Tag{UUID: &t.UUID})
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			err = liberr.Wrap(result.Error)
			return
		} else {
			result = db.First(&tag, model.Tag{Name: t.Name, CategoryID: category.ID})
			if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				err = liberr.Wrap(result.Error)
				return
			}
			err = nil
		}

		tag.Name = t.Name
		tag.UUID = &t.UUID
		tag.CategoryID = category.ID
		result = db.Save(&tag)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	return
}

// Convenience method to find a TagCategory.
func (r *TagCategory) find(db *gorm.DB, conditions ...any) (category *model.TagCategory, found bool, err error) {
	category = &model.TagCategory{}
	result := db.First(category, conditions...)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return
		}
		err = liberr.Wrap(result.Error)
		return
	}
	found = true
	return
}

// Rename a TagCategory by adding a suffix.
func (r *TagCategory) rename(db *gorm.DB, category *model.TagCategory) (err error) {
	suffix := 0
	for {
		suffix++
		newName := fmt.Sprintf("%s (%d)", category.Name, suffix)
		_, found, fErr := r.find(db, "name = ?", newName)
		if fErr != nil {
			err = fErr
			return
		}
		if !found {
			category.Name = newName
			result := db.Save(category)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
			return
		}
	}
}
