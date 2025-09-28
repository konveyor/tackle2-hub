package database

import (
	"database/sql"
	"fmt"
	"time"

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
		db, err = open(
			driver,
			&gorm.Config{
				PrepareStmt:     prod,
				CreateBatchSize: 500,
			})
	} else {
		Settings.DB.MaxConnection = 1
		dsn := fmt.Sprintf("file:%s?_journal=WAL", Settings.DB.Path)
		driver = sqlite.Open(dsn)
		db, err = open(
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
		pdb.SetMaxIdleConns(Settings.DB.MaxConnection / 2)
		pdb.SetConnMaxLifetime(5 * time.Minute)
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

// open used to open GORM with retries.
func open(driver gorm.Dialector, cfg *gorm.Config) (db *gorm.DB, err error) {
	for i := Settings.DB.Retries; i > 0; i-- {
		time.Sleep(time.Second)
		db, err = gorm.Open(driver, cfg)
		if err != nil {
			log.Error(
				err,
				"Database connection failed",
				"retries",
				i)
		} else {
			break
		}
	}
	return
}
