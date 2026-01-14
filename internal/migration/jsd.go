package migration

import (
	"errors"
	"fmt"
	"reflect"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/jsd"
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DocumentMigrator performs migration of
// schema-driven `Document` fields.
type DocumentMigrator struct {
	DB       *gorm.DB
	Client   client.Client
	manager  *jsd.Manager
	versions map[string]Setting
}

// Migrate `Document` fields as needed.
func (dm *DocumentMigrator) Migrate(models []any) (err error) {
	dm.versions = make(map[string]Setting)
	dm.manager = jsd.New(dm.Client)
	err = dm.manager.Load()
	if err != nil {
		return
	}
	err = dm.readSettings()
	if err != nil {
		return
	}
	if dm.skipMigration() {
		Log.Info("jsd: migration skipped.")
		return
	}
	Log.Info("jsd: migration started.")
	err = dm.DB.Transaction(func(tx *gorm.DB) (err error) {
		dm.DB = tx
		for _, m := range dm.hasDocuments(models) {
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
	Log.Info("jsd: migration completed.")
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
	dm.versions = make(map[string]Setting)
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
				setting.Value = Setting{}
			}
		}
		sv := Setting{}
		err = setting.As(&sv)
		if err != nil {
			return
		}
		dm.versions[schema.Name] = sv
	}
	return
}

// updateSettings updates the settings (table) with current schema versions.
func (dm *DocumentMigrator) updateSettings() (err error) {
	dm.versions = make(map[string]Setting)
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
		sv := Setting{
			Digest:  schema.Digest(),
			Version: version,
		}
		dm.versions[schema.Name] = sv
		setting := model.Setting{
			Key:   key,
			Value: sv,
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

// hasDocuments returns the models with `Document` fields.
func (dm *DocumentMigrator) hasDocuments(models []any) (matched []any) {
	for _, m := range models {
		fields := dm.fields(m)
		if len(fields) > 0 {
			matched = append(matched, m)
		}
	}
	return
}

// migrate the `Document` fields as needed.
// Fetch models and migrate them.
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
		err = dm.migrateFields(m)
		if err != nil {
			return
		}
	}
	return
}

// migrateFields migrates the `Document` fields.
func (dm *DocumentMigrator) migrateFields(m any) (err error) {
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
		newCurrent := current.Version
		d.Content, newCurrent, err = schema.Migrate(d.Content, current.Version)
		if err != nil {
			return
		}
		if newCurrent > current.Version {
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

// skipMigration returns true when no schemas has been added or changed.
func (dm *DocumentMigrator) skipMigration() (skip bool) {
	schemas, _ := dm.manager.List()
	for _, schema := range schemas {
		sv := dm.versions[schema.Name]
		digest := schema.Digest()
		if sv.Digest != digest {
			return
		}
	}
	skip = true
	return
}

// key returns the setting (table) key.
func (dm *DocumentMigrator) key(schema string) (key string) {
	key = fmt.Sprintf(".jsd.%s.version", schema)
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

type Setting struct {
	Digest  string `json:"digest"`
	Version int    `json:"version"`
}
