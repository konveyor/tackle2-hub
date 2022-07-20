package database

import (
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

var log = logging.WithName("hub")

//
// Migrate DB to v2.1.0
//
// Applications can be created without a BusinessService in Tackle 2.1.0.
// GORM AutoMigrate doesn't migrate foreign key constraints, so we have to
// explicitly delete and recreate the constraint.
func v210(db *gorm.DB) (err error) {
	var done bool
	done, err = isMigrated(db, Version210)
	if err != nil {
		return
	}
	if done {
		log.V(4).Info("Skipping completed migration.", "version", Version210)
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
	err = updateVersion(db, Version210)
	if err != nil {
		return
	}

	log.V(3).Info("Migration complete.", "version", Version210)
	return
}
