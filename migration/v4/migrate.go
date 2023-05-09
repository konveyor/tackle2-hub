package v4

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v4/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v4")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	// Adding Source column to Fact composite primary key.
	// Altering the primary key requires constructing a new table, so rename the old one,
	// create the new one, copy over the rows, and then drop the old one.
	err = db.Migrator().RenameTable("Fact", "Fact__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Migrator().CreateTable(model.Fact{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	result := db.Exec("INSERT INTO Fact (ApplicationID, Key, Value, Source) SELECT ApplicationID, Key, Value, '' FROM Fact__old;")
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	err = db.Migrator().DropTable("Fact__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
