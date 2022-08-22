package v1

import (
	liberr "github.com/konveyor/controller/pkg/error"
	modelv1 "github.com/konveyor/tackle2-hub/migration/v1/model"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	// More than one row means that this db is a preexisting v2.0.0 deployment,
	// so we should skip this migration.
	result := db.Find(&[]model.Setting{})
	if result.RowsAffected > 1 {
		return
	}
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	result = db.Where("key = ?", ".hub.db.seeded").Delete(&model.Setting{})
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}

	err = db.AutoMigrate(modelv1.All()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	modelv1.Seed(db)

	return
}

func (r Migration) Name() string {
	return "v2.0.0"
}
