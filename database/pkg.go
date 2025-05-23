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

var Log = logr.WithName("db")

var Settings = &settings.Settings

const (
	DSN    = "file:%s?_journal="
	FKsOn  = "&_foreign_keys=yes"
	FKsOff = "&_foreign_keys=no"
)

func init() {
	sql.Register("sqlite3x", &Driver{})
}

// Open and auto-migrate the DB.
// For sqlite3, the default journal mode:
// - DELETE used for NFS.
// - WAL used for non-network filesystems.
func Open(enforceFKs bool) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf(DSN, Settings.DB.Path)
	if Settings.DB.NFS {
		dsn += "DELETE"
	} else {
		dsn += "WAL"
	}
	if enforceFKs {
		dsn += FKsOn
	} else {
		dsn += FKsOff
	}
	dialector := sqlite.Open(dsn).(*sqlite.Dialector)
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
	if Settings.DB.MaxConnection > 0 {
		dbx, nErr := db.DB()
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		dbx.SetMaxOpenConns(Settings.DB.MaxConnection)
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
