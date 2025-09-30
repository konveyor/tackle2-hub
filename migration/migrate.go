package migration

import (
	"errors"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/database/sqlite"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

// Migrate the hub by applying all necessary Migrations.
func Migrate(migrations []Migration) (err error) {
	var db *gorm.DB

	db, err = sqlite.Open(true)
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
	err = setting.As(&v)
	if err != nil {
		return
	}
	var start = v.Version
	if start != 0 && start < MinimumVersion {
		err = errors.New("unsupported database version")
		return
	} else if start >= MinimumVersion {
		start -= MinimumVersion
	}

	if Settings.Hub.Development {
		if start >= len(migrations) {
			Log.Info("Development mode: forcing last migration.")
			start = len(migrations) - 1
		}
	}

	for i := start; i < len(migrations); i++ {
		m := migrations[i]
		ver := i + MinimumVersion + 1

		db, err = sqlite.Open(true)
		if err != nil {
			err = liberr.Wrap(err, "version")
			return
		}

		f := func(db *gorm.DB) (err error) {
			Log.Info("Running migration.", "version", ver)
			err = m.Apply(db)
			if err != nil {
				return
			}
			err = setVersion(db, ver)
			if err != nil {
				return
			}
			return
		}
		err = db.Transaction(f)
		if err != nil {
			err = liberr.Wrap(err, "version", ver)
			return
		}
		err = database.Close(db)
		if err != nil {
			err = liberr.Wrap(err, "version", ver)
			return
		}
	}

	return
}

// Set the version record.
func setVersion(db *gorm.DB, version int) (err error) {
	setting := &model.Setting{Key: VersionKey}
	setting.Value = Version{Version: version}
	result := db.Where("key", VersionKey).Updates(setting)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	return
}
