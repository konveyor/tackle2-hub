package v13

import (
	liberr "github.com/jortel/go-utils/error"
	model2 "github.com/konveyor/tackle2-hub/internal/migration/v13/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.Migrator().DropColumn(&model2.Task{}, "Policy")
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
	return model2.All()
}
