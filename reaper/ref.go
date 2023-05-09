package reaper

import (
	liberr "github.com/jortel/go-utils/error"
	"gorm.io/gorm"
	"reflect"
)

//
// RefCounter provides model inspection for files
// tagged with: ref:<kind>.
type RefCounter struct {
	// DB
	DB *gorm.DB
}

//
// Count find & count references.
func (r *RefCounter) Count(m interface{}, kind string, pk uint) (nRef int64, err error) {
	db := r.DB.Model(m)
	fields := 0
	add := func(ft reflect.StructField) {
		tag, found := ft.Tag.Lookup("ref")
		if found && tag == kind {
			db = db.Or(ft.Name, pk)
			fields++
		}
	}
	var find func(interface{})
	find = func(object interface{}) {
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
			}
		}
	}
	find(m)
	if fields == 0 {
		return
	}
	err = db.Count(&nRef).Error
	if err != nil {
		err = liberr.Wrap(
			err,
			"object",
			reflect.TypeOf(m).Name(),
		)
	}

	return
}
