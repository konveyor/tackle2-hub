package seed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

// RuleSet applies RuleSet seeds.
type RuleSet struct {
	ruleSets []libseed.RuleSet
}

// With collects all the RuleSet seeds.
func (r *RuleSet) With(seed libseed.Seed) (err error) {
	items, err := seed.DecodeItems()
	if err != nil {
		return
	}
	for _, item := range items {
		ruleSet := item.(libseed.RuleSet)
		r.ruleSets = append(r.ruleSets, ruleSet)
	}
	return
}

// Apply seeds the database with RuleSets.
func (r *RuleSet) Apply(db *gorm.DB) (err error) {
	log.Info("Applying RuleSets", "count", len(r.ruleSets))

	ruleSetsByUUID := make(map[string]*model.RuleSet)
	ids := []uint{}

	for i := range r.ruleSets {
		rs := r.ruleSets[i]
		ruleSet, found, fErr := r.find(db, "uuid = ?", rs.UUID)
		if fErr != nil {
			err = fErr
			return
		}
		// model exists and is being renamed
		if found && ruleSet.Name != rs.Name {
			// ensure that the target name is clear
			collision, collides, fErr := r.find(db, "name = ? and id != ?", rs.Name, ruleSet.ID)
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
			ruleSet, found, fErr = r.find(db, "name = ?", rs.Name)
			if fErr != nil {
				err = fErr
				return
			}
			if found && ruleSet.CreateUser != "" {
				err = r.rename(db, ruleSet)
				if err != nil {
					return
				}
				found = false
			}
			if !found {
				ruleSet = &model.RuleSet{}
			}
		}
		ruleSet.Name = rs.Name
		ruleSet.UUID = &rs.UUID
		result := db.Save(ruleSet)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}

		err = r.applyRules(db, ruleSet, rs)
		if err != nil {
			return
		}
		ruleSetsByUUID[rs.UUID] = ruleSet
		ids = append(ids, ruleSet.ID)
	}

	// resolve RuleSet dependency relationships
	for _, rs := range r.ruleSets {
		ruleSet := ruleSetsByUUID[rs.UUID]
		dependsOn := []model.RuleSet{}
		for _, uuid := range rs.Dependencies {
			dep := ruleSetsByUUID[uuid]
			dependsOn = append(dependsOn, *dep)
		}
		ruleSet.DependsOn = dependsOn
		result := db.Save(ruleSet)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}

	value, _ := json.Marshal(ids)
	uiOrder := model.Setting{Key: "ui.ruleset.order", Value: value}
	result := db.Where("key", "ui.ruleset.order").Updates(uiOrder)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// Seed a RuleSet's Rules.
func (r *RuleSet) applyRules(db *gorm.DB, ruleSet *model.RuleSet, rs libseed.RuleSet) (err error) {
	result := db.Delete(&model.Rule{}, "RuleSetID = ?", ruleSet.ID)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	for _, rl := range rs.Rules {
		f, fErr := file(db, rl.Path)
		if fErr != nil {
			err = liberr.Wrap(fErr)
			return
		}
		rule := model.Rule{
			Labels:    rl.Labels(),
			RuleSetID: ruleSet.ID,
			FileID:    &f.ID,
		}
		result = db.Save(&rule)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	return
}

// Create a File model and copy a real file to its path.
func file(db *gorm.DB, filePath string) (file *model.File, err error) {
	file = &model.File{
		Name: path.Base(filePath),
	}
	err = db.Create(&file).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	src, err := os.Open(filePath)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer src.Close()
	dst, err := os.Create(file.Path)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Convenience method to find a RuleSet.
func (r *RuleSet) find(db *gorm.DB, conditions ...any) (rs *model.RuleSet, found bool, err error) {
	rs = &model.RuleSet{}
	result := db.First(rs, conditions...)
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

// Rename a RuleSet by adding a suffix.
func (r *RuleSet) rename(db *gorm.DB, rs *model.RuleSet) (err error) {
	suffix := 0
	for {
		suffix++
		newName := fmt.Sprintf("%s (%d)", rs.Name, suffix)
		_, found, fErr := r.find(db, "name = ?", newName)
		if fErr != nil {
			err = fErr
			return
		}
		if !found {
			rs.Name = newName
			result := db.Save(rs)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
			return
		}
	}
}
