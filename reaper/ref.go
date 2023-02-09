package reaper

import (
	liberr "github.com/konveyor/controller/pkg/error"
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
	mt := reflect.TypeOf(m)
	mv := reflect.ValueOf(m)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	if mv.Kind() != reflect.Struct {
		err = liberr.New("Must be object.")
		return
	}
	db := r.DB.Model(m)
	fields := 0
	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fv := mv.Field(i)
		if !fv.CanSet() {
			continue
		}
		switch fv.Kind() {
		case reflect.Uint:
			tag, found := ft.Tag.Lookup("ref")
			if !found || tag != kind {
				continue
			}
			db = db.Or(ft.Name, pk)
			fields++
		}
	}
	if fields == 0 {
		return
	}
	err = db.Count(&nRef).Error
	if err != nil {
		err = liberr.Wrap(
			err,
			"object",
			mt.Name(),
		)
	}

	return
}

