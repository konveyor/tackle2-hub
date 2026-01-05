package v19

import (
	liberr "github.com/jortel/go-utils/error"
	model2 "github.com/konveyor/tackle2-hub/internal/migration/v19/model"
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
	return model2.All()
}

// This is more complicated than you would think.
// Even if the ApplicationIdentity table is dropped and re-created, the migrator will not
// include the new 'Role' column. As a result, the approach needs to be:
// - create table M (created correctly)
// - populate with the content of ApplicationIdentity.
// - drop table M.
// - drop M_ constraints.
func (r Migration) migrateIdentities(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	if !migrator.HasTable(&model2.ApplicationIdentity{}) {
		return
	}
	if migrator.HasColumn(&model2.ApplicationIdentity{}, "Role") {
		return
	}
	//
	// migrated
	type M struct {
		model2.ApplicationIdentity
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
	//
	// clean up.
	err = migrator.DropTable(&model2.ApplicationIdentity{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.RenameTable("M", "ApplicationIdentity")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropConstraint(&model2.ApplicationIdentity{}, "fk_M_Application")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = migrator.DropConstraint(&model2.ApplicationIdentity{}, "fk_M_Identity")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	//
	// kind=asset no longer used.
	err = db.Model(&model2.Identity{}).
		Where("kind", "asset").
		Update("kind", "source").Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
