package database

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// PK singleton pk sequence.
var PK PkSequence

// PkSequence provides a primary key sequence.
type PkSequence struct {
	mutex sync.Mutex
}

// Load highest key for all models.
func (r *PkSequence) Load(db *gorm.DB, models []any) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, m := range models {
		mt := reflect.TypeOf(m)
		if mt.Kind() == reflect.Ptr {
			mt = mt.Elem()
		}
		kind := strings.ToUpper(mt.Name())
		db = r.session(db)
		q := db.Table(kind)
		q = q.Select("MAX(ID) id")
		cursor, err := q.Rows()
		if err != nil || !cursor.Next() {
			// not a table with id.
			// discarded.
			continue
		}
		id := int64(0)
		err = cursor.Scan(&id)
		_ = cursor.Close()
		if err != nil {
			r.add(db, kind, uint(0))
		} else {
			r.add(db, kind, uint(id))
		}
	}
	return
}

// Next returns the next primary key.
// Updates the LastID.
func (r *PkSequence) Next(db *gorm.DB) (id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	kind := strings.ToUpper(db.Statement.Table)
	m := &model.PK{}
	db = r.session(db)
	err := db.First(m, "Kind", kind).Error
	if err != nil {
		return
	}
	m.LastID++
	id = m.LastID
	err = db.Save(m).Error
	if err != nil {
		panic(err)
	}
	return
}

// Assigned updates LastID (as needed) when explicitly assigned.
func (r *PkSequence) Assigned(db *gorm.DB, id uint) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	kind := strings.ToUpper(db.Statement.Table)
	m := &model.PK{}
	db = r.session(db)
	err := db.First(m, "Kind", kind).Error
	if err != nil {
		return
	}
	if id > m.LastID {
		m.LastID = id
		err = db.Save(m).Error
		if err != nil {
			panic(err)
		}
	}
	return
}

// session returns a new DB with a new session.
func (r *PkSequence) session(in *gorm.DB) (out *gorm.DB) {
	out = &gorm.DB{
		Config: in.Config,
	}
	out.Config.Logger.LogMode(logger.Warn)
	out.Statement = &gorm.Statement{
		DB:       out,
		ConnPool: in.Statement.ConnPool,
		Context:  in.Statement.Context,
		Clauses:  map[string]clause.Clause{},
		Vars:     make([]any, 0, 8),
	}
	return
}

// add the last (higher) id for the kind.
func (r *PkSequence) add(db *gorm.DB, kind string, id uint) {
	m := &model.PK{Kind: kind}
	db = r.session(db)
	err := db.First(m).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(err)
		}
	}
	if m.LastID > id {
		return
	}
	m.LastID = id
	db = r.session(db)
	err = db.Save(m).Error
	if err != nil {
		panic(err)
	}
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
				id, isZero := f.ValueOf(
					statement.Context,
					statement.ReflectValue.Index(i))
				if isZero {
					id = PK.Next(db)
					_ = f.Set(
						statement.Context,
						statement.ReflectValue.Index(i),
						id)

				} else {
					PK.Assigned(db, id.(uint))
				}
				break
			}
		}
	case reflect.Struct:
		for _, f := range schema.Fields {
			if f.Name != "ID" {
				continue
			}
			id, isZero := f.ValueOf(
				statement.Context,
				statement.ReflectValue)
			if isZero {
				id = PK.Next(db)
				_ = f.Set(
					statement.Context,
					statement.ReflectValue,
					id)
			} else {
				PK.Assigned(db, id.(uint))
			}
			break
		}
	default:
		log.Info("[WARN] assignPk: unknown kind.")
	}
}
