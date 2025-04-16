package v2

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v2/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v2")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	constraint := "fk_BusinessService_Applications"
	log.V(4).Info("Dropping constraint.", "constraint", constraint)
	err = db.Migrator().DropConstraint(&model.Application{}, constraint)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	log.V(4).Info("Creating constraint.", "constraint", constraint)
	err = db.Migrator().CreateConstraint(&model.Application{}, constraint)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = seed(db)
	return
}

func (r Migration) Models() []any {
	return model.All()
}
