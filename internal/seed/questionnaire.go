package seed

import (
	"encoding/json"
	"errors"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

// Questionnaire applies Questionnaire seeds.
type Questionnaire struct {
	questionnaires []libseed.Questionnaire
}

// With collects all the Questionnaire seeds.
func (r *Questionnaire) With(seed libseed.Seed) (err error) {
	items, err := seed.DecodeItems()
	if err != nil {
		return
	}
	for _, item := range items {
		r.questionnaires = append(r.questionnaires, item.(libseed.Questionnaire))
	}
	return
}

// Apply seeds the database with Questionnaires.
func (r *Questionnaire) Apply(db *gorm.DB) (err error) {
	log.Info("Applying Questionnaires", "count", len(r.questionnaires))
	for i := range r.questionnaires {
		q := r.questionnaires[i]
		questionnaire, found, fErr := r.find(db, "uuid = ?", q.UUID)
		if fErr != nil {
			err = fErr
			return
		}
		// model exists and is being renamed
		if found && questionnaire.Name != q.Name {
			// ensure that the target name is clear
			collision, collides, fErr := r.find(db, "name = ? and id != ?", q.Name, questionnaire.ID)
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
			questionnaire, found, fErr = r.find(db, "name = ?", q.Name)
			if fErr != nil {
				err = fErr
				return
			}
			if found && questionnaire.CreateUser != "" {
				err = r.rename(db, questionnaire)
				if err != nil {
					return
				}
				found = false
			}
			if !found {
				// only set the required flag on first seed so that
				// we don't e.g. turn it back on when reseeding to
				// fix a typo in the questionnaire or etc.
				questionnaire = &model.Questionnaire{}
				questionnaire.Required = q.Required
			}
		}

		questionnaire.Name = q.Name
		questionnaire.UUID = &q.UUID
		questionnaire.Description = q.Description
		questionnaire.RiskMessages = model.RiskMessages(q.RiskMessages)
		questionnaire.Thresholds = model.Thresholds(q.Thresholds)
		bytes, jErr := json.Marshal(q.Sections)
		if jErr != nil {
			err = liberr.Wrap(jErr)
			return
		}
		err = json.Unmarshal(bytes, &questionnaire.Sections)
		result := db.Save(&questionnaire)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	return
}

// Convenience method to find a Questionnaire.
func (r *Questionnaire) find(db *gorm.DB, conditions ...any) (q *model.Questionnaire, found bool, err error) {
	q = &model.Questionnaire{}
	result := db.First(q, conditions...)
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

// Rename a Questionnaire by adding a suffix.
func (r *Questionnaire) rename(db *gorm.DB, q *model.Questionnaire) (err error) {
	suffix := 0
	for {
		suffix++
		newName := fmt.Sprintf("%s (%d)", q.Name, suffix)
		_, found, fErr := r.find(db, "name = ?", newName)
		if fErr != nil {
			err = fErr
			return
		}
		if !found {
			q.Name = newName
			result := db.Save(q)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
			return
		}
	}
}
