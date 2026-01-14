package v16

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/migration/v16/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = r.dropColumns(db)
	if err != nil {
		return
	}
	err = db.AutoMigrate(r.Models()...)
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
