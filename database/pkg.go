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
	ConnectionString = "file:%s?_journal=WAL&_timeout=5000"
	FKsOn            = "&_foreign_keys=yes"
	FKsOff           = "&_foreign_keys=no"
)

func init() {
	sql.Register("sqlite3x", &Driver{})
}

// Open and automigrate the DB.
func Open(enforceFKs bool) (db *gorm.DB, err error) {
	connStr := fmt.Sprintf(ConnectionString, Settings.DB.Path)
	if enforceFKs {
		connStr += FKsOn
	} else {
		connStr += FKsOff
	}
	dialector := sqlite.Open(connStr).(*sqlite.Dialector)
	dialector.DriverName = "sqlite3x"
	db, err = gorm.Open(
		dialector,
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
