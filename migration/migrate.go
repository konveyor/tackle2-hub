package migration

import (
	"os"
	"path"
	"regexp"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
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
	version, isUpgrade, err := getVersion(db)
	if err != nil {
		return
	}
	err = version.Validate(migrations)
	if err != nil {
		return
	}
	err = database.Close(db)
	if err != nil {
		return
	}

	var beginIndex int
	if !isUpgrade || Settings.Development {
		beginIndex = version.Latest(migrations).Index()
	} else {
		beginIndex = version.Next().Index()
	}
	for i := beginIndex; i < len(migrations); i++ {
		version.With(i)
		m := migrations[i]
		migrate := func() (err error) {
			db, err = database.Open(false)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			defer func() {
				err = database.Close(db)
			}()
			err = db.Transaction(
				func(tx *gorm.DB) (err error) {
					if _, cast := m.(*NopMigration); !cast {
						Log.Info("Running migration.", "version", version.String())
						err = m.Apply(tx)
						if err != nil {
							return
						}
						err = writeSchema(tx, version)
						if err != nil {
							err = liberr.Wrap(err)
							return
						}
					}
					err = setVersion(tx, version)
					if err != nil {
						err = liberr.Wrap(err)
						return
					}
					return
				})
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			return
		}
		err = migrate()
		if err != nil {
			return
		}
	}
	return
}

func getVersion(db *gorm.DB) (v *Version, isUpgrade bool, err error) {
	setting := &model.Setting{}
	result := db.FirstOrCreate(setting, model.Setting{Key: VersionKey})
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	v = &Version{}
	err = setting.As(v)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	isUpgrade = result.RowsAffected == 0
	return
}

// Set the version record.
func setVersion(db *gorm.DB, v *Version) (err error) {
	setting := &model.Setting{Key: VersionKey}
	setting.Value = v
	db = db.Where("key", VersionKey)
	err = db.Updates(setting).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// writeSchema - writes the migrated schema to a file.
func writeSchema(db *gorm.DB, version *Version) (err error) {
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
	f, err := os.Create(path.Join(dir, version.String()))
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
