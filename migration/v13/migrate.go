package v13

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v13/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v12")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
