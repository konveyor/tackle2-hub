package v16

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v16/model"

	"gorm.io/gorm"
)

var log = logr.WithName("migration|v15")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.dropColumns(db)
	return
}

// dropColumns drop unused columns.
// Rank: no longer needed.
// Username: never used.
func (r Migration) dropColumns(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	err = migrator.DropColumn(&model.TagCategory{}, "Rank")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropColumn(&model.TagCategory{}, "Username")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropColumn(&model.Tag{}, "Username")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

func (r Migration) Models() []any {
	return model.All()
}
