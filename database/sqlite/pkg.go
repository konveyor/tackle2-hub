package sqlite

import (
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	Settings = &settings.Settings
	Log      = logr.WithName("sqlite")
)

// Open the DB.
func Open(forMigration bool) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("file:%s?_journal=WAL", Settings.DB.Path)
	if forMigration {
		dsn += "?_foreign_keys=0"
	}
	driver := sqlite.Open(dsn)
	db, err = gorm.Open(
		driver,
		&gorm.Config{
			PrepareStmt:     true,
			CreateBatchSize: 500,
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	pdb, nErr := db.DB()
	if nErr != nil {
		err = liberr.Wrap(nErr)
		return
	}
	pdb.SetMaxOpenConns(1)
	err = db.AutoMigrate(model.PK{}, model.Setting{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = PK.Load(db, []any{model.Setting{}})
	if err != nil {
		return
	}
	err = db.Callback().Create().Before("gorm:before_create").Register("assign-pk", assignPk)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
