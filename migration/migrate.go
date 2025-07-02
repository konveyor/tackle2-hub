package migration

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/jsd"
	"github.com/konveyor/tackle2-hub/migration/json"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	for i := start; i < len(migrations); i++ {
		m := migrations[i]
		ver := i + MinimumVersion + 1

		db, err = database.Open(false)
		if err != nil {
			err = liberr.Wrap(err, "version")
			return
		}

		f := func(db *gorm.DB) (err error) {
			log.Info("Running migration.", "version", ver)
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

	if Settings.Hub.Development {
		log.Info("Running development auto-migration.")
		err = autoMigrate(db, migrations[len(migrations)-1].Models())
		if err != nil {
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

// AutoMigrate the database.
func autoMigrate(db *gorm.DB, models []any) (err error) {
	db, err = database.Open(false)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(models...)
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

// DocumentMigrator performs migration of
// schema-driven `Document` fields.
type DocumentMigrator struct {
	DB       *gorm.DB
	Client   client.Client
	manager  *jsd.Manager
	versions map[string]int
}

// Migrate `Document` fields as needed.
func (dm *DocumentMigrator) Migrate(models []any) (err error) {
	dm.versions = make(map[string]int)
	dm.manager = jsd.New(dm.Client)
	err = dm.manager.Load()
	if err != nil {
		return
	}
	err = dm.readSettings()
	if err != nil {
		return
	}
	err = dm.DB.Transaction(func(tx *gorm.DB) (err error) {
		dm.DB = tx
		for _, m := range dm.withDocuments(models) {
			err = dm.migrate(m)
			if err != nil {
				return
			}
		}
		err = dm.updateSettings()
		if err != nil {
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

// migrate the `Document` fields as needed.
func (dm *DocumentMigrator) migrate(m any) (err error) {
	mt := reflect.TypeOf(m)
	m = reflect.New(mt).Interface()
	db := dm.DB.Model(m)
	db, err = dm.withSelect(db, m)
	if err != nil {
		return
	}
	cursor, err := db.Rows()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	defer func() {
		_ = cursor.Close()
	}()
	for cursor.Next() {
		err = db.ScanRows(cursor, m)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		err = dm.jsdMigrate(m)
		if err != nil {
			return
		}
	}
	return
}

// Fields returns resource `Document` fields.
func (dm *DocumentMigrator) fields(r any) (fields []Field) {
	rt := reflect.TypeOf(r)
	rv := reflect.ValueOf(r)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)
		if !ft.IsExported() {
			continue
		}
		object := fv.Interface()
		switch d := object.(type) {
		case *json.Document:
			fields = append(
				fields,
				Field{
					name:     ft.Name,
					document: d,
				})
		case json.Document:
			fields = append(
				fields,
				Field{
					name:     ft.Name,
					document: &d,
				})
		}
	}
	return
}

// readSettings reads the settings (table) and populates `versions`.
func (dm *DocumentMigrator) readSettings() (err error) {
	dm.versions = make(map[string]int)
	schemas, err := dm.manager.List()
	if err != nil {
		return
	}
	for _, schema := range schemas {
		key := dm.key(schema.Name)
		var setting model.Setting
		err = dm.DB.First(&setting, "key", key).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				err = liberr.Wrap(err)
				return
			} else {
				setting.Value = 0
			}
		}
		version := 0
		err = setting.As(&version)
		if err != nil {
			return
		}
		dm.versions[schema.Name] = version
	}
	return
}

// updateSettings updates the settings (table) with current schema versions.
func (dm *DocumentMigrator) updateSettings() (err error) {
	dm.versions = make(map[string]int)
	schemas, err := dm.manager.List()
	if err != nil {
		return
	}
	for _, schema := range schemas {
		key := dm.key(schema.Name)
		version := len(schema.Versions) - 1
		if version < 0 {
			version = 0
		}
		dm.versions[schema.Name] = version
		setting := model.Setting{
			Key:   key,
			Value: version,
		}
		db := dm.DB.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			})
		err = db.Save(&setting).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// withDocuments returns the models with `Document` fields.
func (dm *DocumentMigrator) withDocuments(models []any) (matched []any) {
	for _, m := range models {
		fields := dm.fields(m)
		if len(fields) > 0 {
			matched = append(matched, m)
		}
	}
	return
}

// jsdMigrate migrates the `Document` fields.
func (dm *DocumentMigrator) jsdMigrate(m any) (err error) {
	var migrated []string
	for _, field := range dm.fields(m) {
		if field.empty() {
			continue
		}
		d := field.document
		schema, nErr := dm.manager.Get(d.Schema)
		if nErr != nil {
			err = nErr
			return
		}
		current := dm.versions[schema.Name]
		newCurrent := current
		d.Content, newCurrent, err = schema.Migrate(d.Content, current)
		if err != nil {
			return
		}
		if newCurrent > current {
			migrated = append(migrated, field.name)
		}
	}
	if len(migrated) > 0 {
		db := dm.DB.Select(migrated)
		err = db.Save(m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

// key returns the setting (table) key.
func (dm *DocumentMigrator) key(schema string) (key string) {
	key = fmt.Sprintf(".jsd.%s.version", schema)
	return
}

// withSelect returns a DB with field names selected.
func (dm *DocumentMigrator) withSelect(in *gorm.DB, m any) (out *gorm.DB, err error) {
	out = in
	names := []string{}
	stmt := &gorm.Statement{DB: in}
	err = stmt.Parse(m)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, field := range stmt.Schema.PrimaryFields {
		names = append(
			names,
			field.Name)
	}
	for _, field := range dm.fields(m) {
		names = append(
			names,
			field.name)
	}
	out = in.Select(names)
	return
}

type Field struct {
	name     string
	document *json.Document
}

func (f *Field) empty() (empty bool) {
	empty = f.document == nil || f.document.Schema == ""
	return
}
