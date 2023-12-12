package v12

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v12/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v11")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
