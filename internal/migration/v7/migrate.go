package v7

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/migration/v6/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	// note: sqlite can't add a unique column, so we add UUID as a optional column,
	// and then mark the column unique and create the index via auto-migrate.
	type TagCategory struct {
		model.TagCategory
		UUID *string
	}

	type Tag struct {
		model.Tag
		UUID *string
	}

	type JobFunction struct {
		model.JobFunction
		UUID *string
	}

	type RuleSet struct {
		model.RuleSet
		UUID *string
	}

	err = db.AutoMigrate(&TagCategory{}, &Tag{}, &JobFunction{}, &RuleSet{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.Delete(&model.RuleSet{}, "CreateUser = ?", "").Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

func (r Migration) Models() []any {
	return model.All()
}
