package database

import (
	"database/sql"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	pg "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var log = logr.WithName("db")

var Settings = &settings.Settings

// Open the DB.
// prod = production (not migration).
func Open(prod bool) (db *gorm.DB, err error) {
	var driver gorm.Dialector
	if prod {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s TimeZone=UTC",
			Settings.DB.Host,
			Settings.DB.Port,
			Settings.DB.User,
			Settings.DB.Password,
			Settings.DB.Name)
		driver = pg.New(pg.Config{
			DSN:                  dsn,
			WithoutQuotingCheck:  true,
			PreferSimpleProtocol: !prod,
		})
		db, err = gorm.Open(
			driver,
			&gorm.Config{
				PrepareStmt:     prod,
				CreateBatchSize: 500,
			})
	} else {
		Settings.DB.MaxConnection = 1
		dsn := fmt.Sprintf("file:%s?_journal=WAL", Settings.DB.Path)
		driver = sqlite.Open(dsn)
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
	}
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if Settings.DB.MaxConnection > 0 {
		pdb, nErr := db.DB()
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		pdb.SetMaxOpenConns(Settings.DB.MaxConnection)
	}
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

// Close the DB.
func Close(db *gorm.DB) (err error) {
	var pdb *sql.DB
	pdb, err = db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = pdb.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
