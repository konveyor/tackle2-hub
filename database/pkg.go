package database

import (
	"database/sql"
	"fmt"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var log = logr.WithName("db")

var Settings = &settings.Settings

const (
	ConnectionString = "file:%s?_journal=WAL"
	FKsOn            = "&_foreign_keys=yes"
	FKsOff           = "&_foreign_keys=no"
)

//
// Open and automigrate the DB.
func Open(enforceFKs bool) (db *gorm.DB, err error) {
	connStr := fmt.Sprintf(ConnectionString, Settings.DB.Path)
	if enforceFKs {
		connStr += FKsOn
	} else {
		connStr += FKsOff
	}
	db, err = gorm.Open(
		sqlite.Open(connStr),
		&gorm.Config{
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
	sqlDB, err := db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	sqlDB.SetMaxOpenConns(1)
	err = db.AutoMigrate(model.Setting{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// Close the DB.
func Close(db *gorm.DB) (err error) {
	var sqlDB *sql.DB
	sqlDB, err = db.DB()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = sqlDB.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
