package v16

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v16/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v15")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []any {
	return model.All()
}
