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
	var fields []string
	var list []map[string]any
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
	db = db.Select(fields)
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(
			err,
			"object",
			reflect.TypeOf(m).Name(),
		)
	}
	for _, ref := range list {
		for _, v := range ref {
			switch n := v.(type) {
			case uint:
				ids[n] = 0
			case *uint:
				if n != nil {
					ids[*n] = 0
				}
			case int64:
				ids[uint(n)] = 0
			}
		}
	}

	return
}
