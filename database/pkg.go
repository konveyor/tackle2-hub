package database

import (
	"encoding/json"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
)

//
// DB constants
const (
	ConnectionString = "file:%s?_foreign_keys=yes"
)

//
// Migration versions
const (
	Version210     = "v2.1.0"
	CurrentVersion = Version210
)

//
// Setup the DB and models.
func Setup() (db *gorm.DB, err error) {
	db, err = open()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	var seeded bool
	seeded, err = isSeeded(db)
	if err != nil {
		return
	}

	dbVersion := &model.Setting{}
	result := db.FirstOrCreate(&dbVersion, model.Setting{Key: api.HubDBVersion})
	if result.Error != nil {
		err = result.Error
		return
	}

	// not seeded and empty version record means this is a fresh db
	// that is already on the latest schema version. Replace empty
	// version record with current version.
	if !seeded && dbVersion.Value == nil {
		err = updateVersion(db, CurrentVersion)
		if err != nil {
			return
		}
	}

	err = migrate(db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	if !seeded {
		err = seed(db)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	return
}

//
// Open and automigrate the DB.
func open() (db *gorm.DB, err error) {
	db, err = gorm.Open(
		sqlite.Open(fmt.Sprintf(ConnectionString, settings.Settings.DB.Path)),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(append(model.All())...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// Check whether the database has already been migrated to the given version.
func isMigrated(db *gorm.DB, version string) (migrated bool, err error) {
	setting := &model.Setting{}
	result := db.Where("key", api.HubDBVersion).Find(setting)
	if result.Error != nil {
		err = result.Error
		return
	}

	if setting.Value.String() >= fmt.Sprintf("\"%s\"", version) {
		migrated = true
	}

	return
}

//
// Perform any necessary DB migrations.
func migrate(db *gorm.DB) (err error) {
	var migrations = []func(*gorm.DB) error{
		v210,
	}

	db.Exec("PRAGMA foreign_keys = OFF")
	for _, m := range migrations {
		err = db.Transaction(m)
		if err != nil {
			return
		}
	}
	db.Exec("PRAGMA foreign_keys = ON")

	return
}

//
// Check whether the database has been seeded.
func isSeeded(db *gorm.DB) (exists bool, err error) {
	result := db.Where("key", api.HubDBSeeded).Find(&model.Setting{})
	if result.Error != nil {
		err = result.Error
		return
	}
	exists = result.RowsAffected > 0
	return
}

//
// Seed the database with the contents of json
// files contained in DB_SEED_PATH.
func seed(db *gorm.DB) (err error) {
	for _, m := range model.All() {
		err = func() (err error) {
			kind := reflect.TypeOf(m).Name()
			fileName := strings.ToLower(kind) + ".json"
			filePath := path.Join(settings.Settings.DB.SeedPath, fileName)
			file, err := os.Open(filePath)
			if err != nil {
				err = nil
				return
			}
			defer file.Close()
			jsonBytes, err := ioutil.ReadAll(file)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}

			var unmarshalled []map[string]interface{}
			err = json.Unmarshal(jsonBytes, &unmarshalled)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			for i := range unmarshalled {
				result := db.Model(&m).Create(unmarshalled[i])
				if result.Error != nil {
					err = result.Error
					return
				}
			}
			return
		}()
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	seeded, _ := json.Marshal(true)
	setting := model.Setting{Key: api.HubDBSeeded, Value: seeded}
	result := db.Create(&setting)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}

	log.V(3).Info("Database seeded.")
	return
}

//
// Update the version record.
func updateVersion(db *gorm.DB, version string) (err error) {
	setting := &model.Setting{Key: api.HubDBVersion}
	value, _ := json.Marshal(version)
	setting.Value = value
	result := db.Where("key", api.HubDBVersion).Updates(setting)
	if result.Error != nil {
		err = result.Error
		return
	}
	return
}
