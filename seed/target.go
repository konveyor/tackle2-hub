package seed

import (
	"encoding/json"
	"errors"
	"fmt"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

const UITargetOrder = "ui.target.order"

//
// Target applies Target seeds.
type Target struct {
	targets []libseed.Target
}

//
// With collects all the Target seeds.
func (r *Target) With(seed libseed.Seed) (err error) {
	items, err := seed.DecodeItems()
	if err != nil {
		return
	}
	for _, item := range items {
		r.targets = append(r.targets, item.(libseed.Target))
	}
	return
}

//
// Apply seeds the database with JobFunctions.
func (r *Target) Apply(db *gorm.DB) (err error) {
	log.Info("Applying Targets", "count", len(r.targets))

	for i := range r.targets {
		t := r.targets[i]
		target, found, fErr := r.find(db, "uuid = ?", t.UUID)
		if fErr != nil {
			err = fErr
			return
		}
		// model exists and is being renamed
		if found && target.Name != t.Name {
			// ensure that the target name is clear
			collision, collides, fErr := r.find(db, "name = ? and id != ?", t.Name, target.ID)
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
			target, found, fErr = r.find(db, "name = ?", t.Name)
			if fErr != nil {
				err = fErr
				return
			}
			if found && target.CreateUser != "" {
				err = r.rename(db, target)
				if err != nil {
					return
				}
				found = false
			}
			if !found {
				target = &model.Target{}
			}
		}

		f, fErr := file(db, t.Image())
		if fErr != nil {
			err = liberr.Wrap(fErr)
			return
		}
		labels, _ := json.Marshal(t.Labels)

		target.UUID = &t.UUID
		target.Name = t.Name
		target.Description = t.Description
		target.Choice = t.Choice
		target.ImageID = f.ID
		target.Labels = labels
		result := db.Save(&target)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}

	err = r.reorder(db)
	if err != nil {
		return
	}
	return
}

//
// reorder updates the value of the ui.target.order setting
// to add any missing target ids. (namely, newly added targets.)
func (r *Target) reorder(db *gorm.DB) (err error) {
	targets := []model.Target{}
	result := db.Find(&targets)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}
	s := model.Setting{}
	result = db.First(&s, "key", UITargetOrder)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}
	ordering := []uint{}
	_ = s.As(&ordering)
	known := make(map[uint]bool)
	for _, id := range ordering {
		known[id] = true
	}
	for _, t := range targets {
		if !known[t.ID] {
			ordering = append(ordering, t.ID)
		}
	}
	err = s.With(ordering)
	if err != nil {
		return
	}
	result = db.Where("key", UITargetOrder).Updates(s)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// Convenience method to find a Target.
func (r *Target) find(db *gorm.DB, conditions ...interface{}) (t *model.Target, found bool, err error) {
	t = &model.Target{}
	result := db.First(t, conditions...)
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

//
// Rename a Target by adding a suffix.
func (r *Target) rename(db *gorm.DB, t *model.Target) (err error) {
	suffix := 0
	for {
		suffix++
		newName := fmt.Sprintf("%s (%d)", t.Name, suffix)
		_, found, fErr := r.find(db, "name = ?", newName)
		if fErr != nil {
			err = fErr
			return
		}
		if !found {
			t.Name = newName
			result := db.Save(t)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
			return
		}
	}
}
