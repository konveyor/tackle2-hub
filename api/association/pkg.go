package association

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// New returns an association.
func New(db *gorm.DB, name string) *Association {
	return &Association{
		db:   db,
		name: name,
	}
}

type AssociationError struct {
	Reason string
}

func (r *AssociationError) Error() string {
	return r.Reason
}

func (r *AssociationError) Is(err error) (matched bool) {
	_, matched = err.(*AssociationError)
	return
}

// Association provides association management.
// The gorm.Association.Replace() does not support Omit(clause.Associations).
// As a result, associated models that do not exist are created.
// This behavior is not desirable.
type Association struct {
	db    *gorm.DB
	name  string
	owner bool
}

// DB set the db.
func (r *Association) DB(v *gorm.DB) *Association {
	r.db = v
	return r
}

// Owner sets the owner flag.
// Relation: many2many:
//
//	No effect.
//
// Relation: hasMany:
//
//	Models no longer associated:
//	  When (true): deleted.
//	  When (false): FK=NULL.
func (r *Association) Owner(v bool) *Association {
	r.owner = v
	return r
}

// Name set the association name.
func (r *Association) Name(name string) *Association {
	r.name = name
	return r
}

// Replace associations.
// Supported:
//   - many2many
//   - hasMany
func (a *Association) Replace(m any, related any) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = &AssociationError{
				Reason: r.(error).Error(),
			}
		}
	}()
	db := a.db
	db = db.Model(m)
	related, err = a.asList(related)
	if err != nil {
		return
	}
	relatedIDs, err := a.relatedIDs(db, related.([]any))
	if err != nil {
		return
	}
	stmt := &gorm.Statement{DB: db}
	err = stmt.Parse(m)
	if err != nil {
		return
	}
	relation, found := stmt.Schema.Relationships.Relations[a.name]
	if !found {
		err = &AssociationError{
			Reason: "Association not found.",
		}
		return
	}
	pkField := stmt.Schema.PrioritizedPrimaryField
	if pkField == nil {
		err = &AssociationError{
			Reason: "PK (field) not found.",
		}
		return
	}
	mv := reflect.ValueOf(m)
	pkv := pkField.ReflectValueOf(stmt.Context, mv)
	pk := pkv.Interface()
	switch relation.Type {
	case schema.Many2Many:
		db := a.db
		joinTable := relation.JoinTable
		if joinTable == nil {
			err = &AssociationError{
				Reason: "Join refTable not found.",
			}
			return
		}
		oneRef := relation.References[0]
		oneField := oneRef.ForeignKey.DBName
		manyRef := relation.References[1]
		manyField := manyRef.ForeignKey.DBName
		q := fmt.Sprintf(
			"DELETE FROM %s WHERE %s = ?",
			joinTable.Table,
			oneField)
		err = db.Exec(q, pk).Error
		if err != nil {
			return
		}
		for _, id := range relatedIDs {
			q := fmt.Sprintf(
				"INSERT INTO %s (%s, %s) VALUES (?, ?)",
				joinTable.Table,
				oneField,
				manyField,
			)
			err = db.Exec(q, pk, id).Error
			if err != nil {
				return
			}
		}
	case schema.HasMany:
		db := a.db.Omit(clause.Associations)
		ref := relation.References[0]
		fkField := ref.ForeignKey.DBName
		refField := ref.PrimaryKey.DBName
		refTable := relation.FieldSchema.Table
		var q string
		if a.owner {
			q = fmt.Sprintf(
				"DELETE FROM %s WHERE %s = ?",
				refTable,
				fkField,
			)
		} else {
			q = fmt.Sprintf(
				"UPDATE %s SET %s = NULL WHERE %s = ?",
				refTable,
				fkField,
				fkField,
			)
		}
		err = db.Exec(q, pk).Error
		if err != nil {
			return
		}
		if a.owner {
			for i := range related.([]any) {
				m := related.([]any)[i]
				err = a.set(m, fkField, pk)
				if err != nil {
					return
				}
				err = db.Create(m).Error
				if err != nil {
					return
				}
			}
		} else {
			q = fmt.Sprintf(
				"UPDATE %s SET %s = ? WHERE %s IN (?)",
				refTable,
				fkField,
				refField,
			)
			err = db.Exec(q, pk, relatedIDs).Error
			if err != nil {
				return
			}
		}
	default:
		err = &AssociationError{
			Reason: "Association (kind) not supported.",
		}
		return
	}
	return
}

// set field value.
func (a *Association) set(m any, field string, value any) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			if rErr, cast := r.(error); cast {
				err = rErr
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	mv := reflect.ValueOf(m)
	f := mv.Elem()
	fv := f.FieldByName(field)
	vv := reflect.ValueOf(value)
	fv.Set(vv)
	return
}

// asList returns the any as []*model.
func (a *Association) asList(in any) (out []any, err error) {
	mv := reflect.ValueOf(in)
	if mv.Kind() != reflect.Slice {
		err = &AssociationError{
			Reason: "Must be SLICE.",
		}
		return
	}
	for i := 0; i < mv.Len(); i++ {
		v := mv.Index(i)
		if v.Kind() != reflect.Ptr {
			v = v.Addr()
		}
		m := v.Interface()
		out = append(
			out,
			m)
	}
	return
}

// relatedIDs returns a list of models IDs.
func (a *Association) relatedIDs(db *gorm.DB, related []any) (ids []uint, err error) {
	for _, m := range related {
		id, err := a.ID(db, m)
		if err == nil {
			ids = append(ids, id)
		} else {
			break
		}
	}
	return
}

// ID returns a model ID.
func (a *Association) ID(db *gorm.DB, m any) (id uint, err error) {
	stmt := &gorm.Statement{DB: db}
	err = stmt.Parse(m)
	if err != nil {
		return
	}
	field := stmt.Schema.PrioritizedPrimaryField
	if field == nil {
		err = &AssociationError{
			Reason: "PK (field) not found.",
		}
		return
	}
	value := field.ReflectValueOf(
		stmt.Context,
		reflect.ValueOf(m))
	switch value.Kind() {
	case reflect.Uint,
		reflect.Uint64:
		id = uint(value.Uint())
	default:
		err = &AssociationError{
			Reason: "PK (type) not supported.",
		}
		return
	}
	return
}
