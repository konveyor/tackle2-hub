package reaper

import (
	"fmt"
	"reflect"

	liberr "github.com/jortel/go-utils/error"
	"gorm.io/gorm"
)

// RefFinder provides model inspection for files
// tagged with:
//
//	ref:<kind>
//	[]ref:<kind>
type RefFinder struct {
	// DB
	DB *gorm.DB
}

// Find returns a map of all references for the model and kind.
func (r *RefFinder) Find(m any, kind string, ids map[uint]byte) (err error) {
	var nfields []string
	var jfields []string
	add := func(ft reflect.StructField) {
		tag, found := ft.Tag.Lookup("ref")
		if found && tag == kind {
			nfields = append(
				nfields,
				ft.Name)
			return
		}
		if found && tag == "[]"+kind {
			jfields = append(
				jfields,
				ft.Name)
		}
	}
	var find func(any)
	find = func(object any) {
		mt := reflect.TypeOf(object)
		mv := reflect.ValueOf(object)
		if mt.Kind() == reflect.Ptr {
			mt = mt.Elem()
			mv = mv.Elem()
		}
		if mv.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			if !ft.IsExported() {
				continue
			}
			switch fv.Kind() {
			case reflect.Struct:
				find(fv.Interface())
			case reflect.Ptr:
				inst := fv.Interface()
				switch inst.(type) {
				case *uint:
					add(ft)
				default:
					find(fv.Interface())
				}
			case reflect.Uint:
				add(ft)
			case reflect.Slice:
				add(ft)
			default:
			}
		}
	}
	find(m)
	if len(nfields)+len(jfields) == 0 {
		return
	}
	db := r.DB.Model(m)
	if Log.V(1).Enabled() {
		db = db.Debug()
	}
	var fields []string
	for i := range nfields {
		fields = append(fields, nfields[i])
	}
	for i := range jfields {
		fields = append(
			fields,
			fmt.Sprintf(
				"json_extract(j%d.value,'$.id')",
				i))
		db = db.Joins(
			fmt.Sprintf(
				",json_each(%s) j%d",
				jfields[i],
				i))
	}
	ref := make(map[string]any)
	db = db.Select(fields)
	db = db.Debug()
	cursor, err := db.Rows()
	if err != nil {
		err = liberr.Wrap(
			err,
			"object",
			reflect.TypeOf(m).Name(),
		)
	}
	defer func() {
		_ = cursor.Close()
	}()
	for cursor.Next() {
		err = db.ScanRows(cursor, &ref)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		for _, v := range ref {
			n := r.id(v)
			if n > 0 {
				ids[n] = 0
			}
		}
	}

	return
}

// id returns the uint (id).
// Zero(0) indicates the type could not be converted.
func (r *RefFinder) id(v any) (id uint) {
	defer func() {
		recover()
	}()
	switch n := v.(type) {
	case *uint:
		if n != nil {
			id = *n
		}
	case int64:
		id = uint(n)
	case *int64:
		if n != nil {
			id = uint(*n)
		}
	case *any:
		if n != nil {
			id = r.id(*n)
		}
	default:
		fmt.Printf("unknown: %T\n", n)
	}
	return
}
