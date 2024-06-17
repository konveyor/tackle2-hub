package v13

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v13/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v13")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.Migrator().DropColumn(&model.Task{}, "Policy")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	return
}

func (r Migration) Models() []any {
	return model.All()
}
