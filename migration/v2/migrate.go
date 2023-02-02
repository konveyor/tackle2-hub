package v2

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/migration/v2/model"
	"gorm.io/gorm"
)

var log = logging.WithName("migration|v2")

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
		return
	}
	log.V(4).Info("Creating constraint.", "constraint", constraint)
	err = db.Migrator().CreateConstraint(&model.Application{}, constraint)
	if err != nil {
		return
	}

	model.Seed(db)

	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
