package seed

import (
	"container/list"
	"errors"
	"fmt"
	"slices"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

const UITargetOrder = "ui.target.order"

// Target applies Target seeds.
type Target struct {
	targets []libseed.Target
}

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

// Apply seeds the database with JobFunctions.
func (r *Target) Apply(db *gorm.DB) (err error) {
	log.Info("Applying Targets", "count", len(r.targets))

	var seedIds []uint
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

		target.UUID = &t.UUID
		target.Name = t.Name
		target.Description = t.Description
		target.Provider = t.Provider
		target.Choice = t.Choice
		target.ImageID = f.ID
		target.Labels = []model.TargetLabel{}
		for _, l := range t.Labels {
			target.Labels = append(target.Labels, model.TargetLabel(l))
		}
		result := db.Save(&target)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
		seedIds = append(seedIds, target.ID)
	}

	err = r.deleteUnwanted(db)
	if err != nil {
		return
	}

	err = r.reorder(db, seedIds)
	if err != nil {
		return
	}
	return
}

// deleteUnwanted deletes targets with a UUID not found
// in the set of seeded targets.
func (r *Target) deleteUnwanted(db *gorm.DB) (err error) {
	var found []*model.Target
	err = db.Find(&found).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	wanted := make(map[string]byte)
	for _, t := range r.targets {
		wanted[t.UUID] = 0
	}
	for _, target := range found {
		if target.UUID == nil {
			continue
		}
		if _, found := wanted[*target.UUID]; found {
			continue
		}
		err = db.Delete(target).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// reorder updates the value of the ui.target.order setting
// to add any missing target ids. (namely, newly added targets.)
func (r *Target) reorder(db *gorm.DB, seedIds []uint) (err error) {
	targets := []model.Target{}
	result := db.Find(&targets)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}
	var targetIds []uint
	for _, t := range targets {
		targetIds = append(targetIds, t.ID)
	}

	s := model.Setting{}
	result = db.First(&s, "key", UITargetOrder)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}
	userOrder := []uint{}
	_ = s.As(&userOrder)
	s.Value = r.merge(userOrder, seedIds, targetIds)

	result = db.Where("key", UITargetOrder).Updates(s)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// merge new targets into the user's custom target order.
//
//	params:
//	  userOrder: slice of target IDs in the user's desired order
//	  seedOrder: slice of target IDs in seedfile order
//	  ids: slice of ids of all the targets in the DB
func (r *Target) merge(userOrder []uint, seedOrder []uint, ids []uint) (mergedOrder []uint) {
	ll := list.New()
	known := make(map[uint]*list.Element)
	for _, id := range userOrder {
		if slices.Contains(ids, id) {
			known[id] = ll.PushBack(id)
		}
	}
	for i, id := range seedOrder {
		if _, found := known[id]; found {
			continue
		}
		if i == 0 {
			known[id] = ll.PushFront(id)
		} else {
			known[id] = ll.InsertAfter(id, known[seedOrder[i-1]])
		}
	}

	for _, id := range ids {
		if _, found := known[id]; found {
			continue
		}
		ll.PushBack(id)
	}

	for ll.Len() > 0 {
		e := ll.Front()
		mergedOrder = append(mergedOrder, e.Value.(uint))
		ll.Remove(e)
	}

	return
}

// Convenience method to find a Target.
func (r *Target) find(db *gorm.DB, conditions ...any) (t *model.Target, found bool, err error) {
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
