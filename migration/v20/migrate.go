package v20

import (
	"reflect"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/migration/v20/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v20")

type Migration struct{}

func (r Migration) Models() []any {
	return model.All()
}

func (r Migration) Apply(sqlite *gorm.DB) (err error) {
	db, err := database.Open(true)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = r.migrateData(
		db,
		sqlite,
		[]model.PK{},
		[]model.Setting{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

func (r Migration) migrateData(db, sqlite *gorm.DB, collections ...any) (err error) {
	for _, list := range collections {
		err = sqlite.Find(&list).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if reflect.ValueOf(list).Len() == 0 {
			return
		}
		err = db.Create(&list).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}
