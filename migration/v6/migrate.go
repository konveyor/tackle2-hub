package v6

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v6/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v5")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	m := db.Migrator()
	err = m.DropIndex(model.TechDependency{}, "DepA")
	if err != nil {
		return
	}
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
