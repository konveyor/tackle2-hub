package v3

import (
	"encoding/json"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	v2 "github.com/konveyor/tackle2-hub/migration/v2/model"
	"github.com/konveyor/tackle2-hub/migration/v3/model"
	"github.com/konveyor/tackle2-hub/migration/v3/seed"
	"gorm.io/gorm"
)

var log = logging.WithName("migration|v3")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = r.factMigration(db)
	if err != nil {
		return
	}

	err = db.Migrator().RenameTable(model.TagType{}, model.TagCategory{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.Migrator().RenameColumn(model.Tag{}, "TagTypeID", "CategoryID")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.Migrator().RenameColumn(model.ImportTag{}, "TagType", "Category")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	// Create tables for Trackers and Tickets
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	//
	// Seed.
	seed.Seed(db)

	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}

//
// factMigration migrates Application.Facts.
// This involves changing the Facts type from JSON which maps to
// a column in the DB to an ORM virtual field. This, and the data
// migration both require the v2 model.
func (r Migration) factMigration(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	list := []v2.Application{}
	result := db.Find(&list)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	err = migrator.AutoMigrate(&model.Fact{})
	if err != nil {
		return
	}
	for _, m := range list {
		d := map[string]interface{}{}
		_ = json.Unmarshal(m.Facts, &d)
		for k, v := range d {
			jv, _ := json.Marshal(v)
			fact := &model.Fact{}
			fact.ApplicationID = m.ID
			fact.Key = k
			fact.Value = jv
			result := db.Create(fact)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
		}
	}
	err = migrator.DropColumn(&v2.Application{}, "Facts")
	if err != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	return
}
