package database

import (
	"database/sql"
	"fmt"
	"reflect"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/generated"
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
	err = generated.PK.Load(db, []any{model.Setting{}})
	if err != nil {
		return
	}
	err = db.Callback().Create().Before("gorm:before_create").Register(
		"pk",
		func(db *gorm.DB) {
			statement := db.Statement
			schema := statement.Schema
			if schema == nil {
				return
			}
			switch statement.ReflectValue.Kind() {
			case reflect.Slice,
				reflect.Array:
				for i := 0; i < statement.ReflectValue.Len(); i++ {
					for _, f := range schema.Fields {
						if f.Name != "ID" {
							continue
						}
						_, isZero := f.ValueOf(
							statement.Context,
							statement.ReflectValue.Index(i))
						if isZero {
							id := generated.PK.Next(db.Statement.Table)
							_ = f.Set(
								statement.Context,
								statement.ReflectValue.Index(i),
								id)

						}
						break
					}
				}
			case reflect.Struct:
				for _, f := range schema.Fields {
					if f.Name != "ID" {
						continue
					}
					_, isZero := f.ValueOf(
						statement.Context,
						statement.ReflectValue)
					if isZero {
						id := generated.PK.Next(db.Statement.Table)
						_ = f.Set(
							statement.Context,
							statement.ReflectValue,
							id)

					}
					break
				}
			}
		})
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
