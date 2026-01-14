package seed

import (
	"errors"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

// JobFunction applies JobFunction seeds.
type JobFunction struct {
	jobFunctions []libseed.JobFunction
}

// With collects all the JobFunction seeds.
func (r *JobFunction) With(seed libseed.Seed) (err error) {
	items, err := seed.DecodeItems()
	if err != nil {
		return
	}
	for _, item := range items {
		r.jobFunctions = append(r.jobFunctions, item.(libseed.JobFunction))
	}
	return
}

// Apply seeds the database with JobFunctions.
func (r *JobFunction) Apply(db *gorm.DB) (err error) {
	log.Info("Applying JobFunctions", "count", len(r.jobFunctions))
	for i := range r.jobFunctions {
		jf := r.jobFunctions[i]
		jobFunction, found, fErr := r.find(db, "uuid = ?", jf.UUID)
		if fErr != nil {
			err = fErr
			return
		}
		// model exists and is being renamed
		if found && jobFunction.Name != jf.Name {
			// ensure that the target name is clear
			collision, collides, fErr := r.find(db, "name = ? and id != ?", jf.Name, jobFunction.ID)
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
			jobFunction, found, fErr = r.find(db, "name = ?", jf.Name)
			if fErr != nil {
				err = fErr
				return
			}
			if found && jobFunction.CreateUser != "" {
				err = r.rename(db, jobFunction)
				if err != nil {
					return
				}
				found = false
			}
			if !found {
				jobFunction = &model.JobFunction{}
			}
		}

		jobFunction.Name = jf.Name
		jobFunction.UUID = &jf.UUID
		result := db.Save(&jobFunction)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	return
}

// Convenience method to find a JobFunction.
func (r *JobFunction) find(db *gorm.DB, conditions ...any) (jf *model.JobFunction, found bool, err error) {
	jf = &model.JobFunction{}
	result := db.First(jf, conditions...)
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

// Rename a JobFunction by adding a suffix.
func (r *JobFunction) rename(db *gorm.DB, jf *model.JobFunction) (err error) {
	suffix := 0
	for {
		suffix++
		newName := fmt.Sprintf("%s (%d)", jf.Name, suffix)
		_, found, fErr := r.find(db, "name = ?", newName)
		if fErr != nil {
			err = fErr
			return
		}
		if !found {
			jf.Name = newName
			result := db.Save(jf)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
			return
		}
	}
}
