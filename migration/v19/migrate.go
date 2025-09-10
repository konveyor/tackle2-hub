package v19

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/migration/v19/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = r.migrateIdentities(db)
	if err != nil {
		return
	}
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []any {
	return model.All()
}

func (r Migration) migrateIdentities(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	err = migrator.RenameTable("ApplicationIdentity", "saved")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(&model.ApplicationIdentity{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	type M struct {
		model.ApplicationIdentity
		Kind string
	}
	var saved []M
	db2 := db.Table("saved")
	db2 = db2.Joins("INNER JOIN Identity id ON  id.ID = saved.ApplicationIdentity.IdentityID")
	err = db2.Table("appIdSaved").Find(&saved).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	db3 := db.Clauses(clause.OnConflict{DoNothing: true})
	for _, m := range saved {
		m.Role = m.Kind
		err = db3.Create(&m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	err = migrator.DropTable("saved")
	return
}
