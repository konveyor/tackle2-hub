package migration

import (
	"encoding/json"
	"errors"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

//
// Migrate the hub by applying all necessary Migrations.
func Migrate(migrations []Migration) (err error) {
	var db *gorm.DB

	db, err = database.Open(false)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = database.Close(db)
		}
	}()

	setting := &model.Setting{}
	result := db.FirstOrCreate(setting, model.Setting{Key: VersionKey})
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}

	err = database.Close(db)
	if err != nil {
		return
	}

	var v Version
	if setting.Value != nil {
		err = json.Unmarshal(setting.Value, &v)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	var start = v.Version
	if start == 0 {
		start = SupportedFrom
	} else if start < SupportedFrom {
		err = errors.New("unsupported database version")
		log.Error(err, "Unable to migrate database.", "version", v.Version)
		return
	}

	// Version is the index of the last successful migration,
	// so we want to start iteration at the next index.
	migrations = append([]Migration{nil}, migrations...)
	for i := start + 1; i < len(migrations); i++ {
		m := migrations[i]

		db, err = database.Open(false)
		if err != nil {
			err = liberr.Wrap(err, "version", i)
			return
		}

		f := func(db *gorm.DB) (err error) {
			log.Info("Running migration.", "version", i)
			err = m.Apply(db)
			if err != nil {
				return
			}
			err = setVersion(db, i)
			if err != nil {
				return
			}
			return
		}
		err = db.Transaction(f)
		if err != nil {
			err = liberr.Wrap(err, "version", i)
			return
		}

		err = database.Close(db)
		if err != nil {
			err = liberr.Wrap(err, "version", i)
			return
		}
	}

	if Settings.Hub.Development {
		log.Info("Running development auto-migration.")
		err = autoMigrate(db, migrations[len(migrations)-1].Models())
		if err != nil {
			return
		}
	}

	return
}

//
// Set the version record.
func setVersion(db *gorm.DB, version int) (err error) {
	setting := &model.Setting{Key: VersionKey}
	v := Version{Version: version}
	value, _ := json.Marshal(v)
	setting.Value = value
	result := db.Where("key", VersionKey).Updates(setting)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	return
}

//
// AutoMigrate the database.
func autoMigrate(db *gorm.DB, models []interface{}) (err error) {
	db, err = database.Open(false)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(models)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = database.Close(db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
