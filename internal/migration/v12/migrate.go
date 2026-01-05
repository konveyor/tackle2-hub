package v12

import (
	"github.com/konveyor/tackle2-hub/internal/migration/v12/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []any {
	return model.All()
}
