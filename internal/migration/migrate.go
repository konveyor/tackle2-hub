package migration

import (
	"errors"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"gorm.io/gorm"
)

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

		db, err = database.Open(false)
		if err != nil {
			err = liberr.Wrap(err, "version")
			return
		}
		if Settings.Log.Migration > 0 {
			db = db.Debug()
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
		err = writeSchema(db, ver)
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

// writeSchema - writes the migrated schema to a file.
func writeSchema(db *gorm.DB, version int) (err error) {
	var list []struct {
		Type     string `gorm:"column:type"`
		Name     string `gorm:"column:name"`
		Table    string `gorm:"column:tbl_name"`
		RootPage int    `gorm:"column:rootpage"`
		SQL      string `gorm:"column:sql"`
	}
	db = db.Table("sqlite_schema")
	db = db.Order("1, 2")
	err = db.Find(&list).Error
	if err != nil {
		return
	}
	dir := path.Join(
		path.Dir(Settings.Hub.DB.Path),
		"migration")
	err = nas.MkDir(dir, 0755)
	f, err := os.Create(path.Join(dir, strconv.Itoa(version)))
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	pattern := regexp.MustCompile(`[,()]`)
	SQL := func(in string) (out string) {
		indent := "\n    "
		for {
			m := pattern.FindStringIndex(in)
			if m == nil {
				out += in
				break
			}
			out += indent
			out += in[:m[0]]
			out += indent
			out += in[m[0]:m[1]]
			in = in[m[1]:]
		}
		return
	}
	for _, m := range list {
		s := strings.Join([]string{
			m.Type,
			m.Name,
			m.Table,
			SQL(m.SQL),
		}, "|")
		_, err = f.WriteString(s + "\n")
		if err != nil {
			return
		}
	}
	return
}
