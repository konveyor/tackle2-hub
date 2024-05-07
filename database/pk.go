package database

import (
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// PK singleton pk sequence.
var PK PkSequence

// PkSequence provides a primary key sequence.
type PkSequence struct {
	keySet map[string]uint
	mutex  sync.Mutex
}

// Load highest key for all models.
func (r *PkSequence) Load(db *gorm.DB, models []any) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.keySet = make(map[string]uint)
	for _, m := range models {
		mt := reflect.TypeOf(m)
		list := make([]map[string]any, 0)
		kind := strings.ToUpper(mt.Name())
		q := db.Table(kind)
		q = q.Select("ID")
		err = q.Find(&list).Error
		if err != nil {
			r.keySet[kind] = uint(0)
			err = nil
			continue
		}
		for _, m := range list {
			v := m["ID"]
			switch id := v.(type) {
			case uint:
				r.keySet[kind] = id
			case uint8:
				r.keySet[kind] = uint(id)
			case uint16:
				r.keySet[kind] = uint(id)
			case uint32:
				r.keySet[kind] = uint(id)
			case uint64:
				r.keySet[kind] = uint(id)
			case int:
				r.keySet[kind] = uint(id)
			case int8:
				r.keySet[kind] = uint(id)
			case int16:
				r.keySet[kind] = uint(id)
			case int32:
				r.keySet[kind] = uint(id)
			case int64:
				r.keySet[kind] = uint(id)
			}
		}
	}
	return
}

// Next returns the next primary key.
func (r *PkSequence) Next(kind string) (id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	kind = strings.ToUpper(kind)
	id, found := r.keySet[kind]
	if found {
		id++
		r.keySet[kind] = id
	}
	return
}

// assignPk assigns PK as needed.
func assignPk(db *gorm.DB) {
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
					id := PK.Next(db.Statement.Table)
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
				id := PK.Next(db.Statement.Table)
				_ = f.Set(
					statement.Context,
					statement.ReflectValue,
					id)

			}
			break
		}
	default:
		log.Info("[WARN] assignPk: unknown kind.")
	}
}
