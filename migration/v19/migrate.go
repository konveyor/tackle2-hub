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
	db = db.Debug()
	migrator := db.Migrator()
	type M struct {
		model.ApplicationIdentity
	}
	err = db.AutoMigrate(&M{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	var saved []M
	db2 := db.Table("ApplicationIdentity")
	db2 = db2.Select(
		"IdentityID",
		"ApplicationID",
		"Kind Role")
	db2 = db2.Joins("INNER JOIN Identity id ON  id.ID = IdentityID")
	err = db2.Find(&saved).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	db3 := db.Omit(clause.Associations)
	db3 = db3.Clauses(clause.OnConflict{DoNothing: true})
	for _, m := range saved {
		m2 := &M{}
		m2.IdentityID = m.IdentityID
		m2.ApplicationID = m.ApplicationID
		m2.Role = m.Role
		err = db3.Create(m2).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	err = migrator.DropTable(&model.ApplicationIdentity{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.RenameTable("M", "ApplicationIdentity")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropConstraint(&model.ApplicationIdentity{}, "fk_M_Application")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropConstraint(&model.ApplicationIdentity{}, "fk_M_Identity")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
