package v9

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v9/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v9")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {

	type Review struct {
		model.Review
		ArchetypeID *uint
	}

	err = db.AutoMigrate(&Review{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []any {
	return model.All()
}
