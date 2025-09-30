package v20

import (
	"fmt"
	"reflect"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/database/postgres"
	v19 "github.com/konveyor/tackle2-hub/migration/v19/model"
	"github.com/konveyor/tackle2-hub/migration/v20/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Log = logr.WithName("migration|v20")

type Migration struct{}

func (r Migration) Models() []any {
	return model.All()
}

func (r Migration) Apply(sqlite *gorm.DB) (err error) {
	db, err := postgres.Open(true)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, p := range []Pair{
		{mA: v19.Setting{}, mB: model.Setting{}},
		{mA: v19.Proxy{}, mB: model.Proxy{}},
	} {
		err = r.migratePair(sqlite, db, p)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

func (r Migration) migratePair(sqlite, db *gorm.DB, p Pair) (err error) {
	vA := reflect.ValueOf(p.mA)
	vB := reflect.ValueOf(p.mB)
	if vA.Kind() == reflect.Ptr {
		vA = vA.Elem()
	}
	if vB.Kind() == reflect.Ptr {
		vB = vB.Elem()
	}
	ptA := reflect.PointerTo(vA.Type())
	stA := reflect.SliceOf(ptA)
	svA := reflect.New(stA)
	sA := svA.Interface()
	err = sqlite.Find(sA).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	count := svA.Elem().Len()
	if count == 0 {
		return
	}
	ptB := reflect.PointerTo(vB.Type())
	stB := reflect.SliceOf(ptB)
	svB := reflect.New(stB)
	sB := svB.Elem()
	for i := 0; i < count; i++ {
		mA := svA.Elem().Index(i).Interface()
		mB := reflect.New(ptB.Elem()).Interface()
		err = r.migrate(mA, mB)
		if err == nil {
			sB.Set(reflect.Append(sB, reflect.ValueOf(mB)))
		} else {
			err = liberr.Wrap(err)
			return
		}
	}
	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(sB.Interface()).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

func (r Migration) migrate(mA, mB any) (err error) {
	vA := reflect.ValueOf(mA)
	vB := reflect.ValueOf(mB)
	if vA.Kind() == reflect.Ptr {
		vA = vA.Elem()
	} else {
		err = fmt.Errorf("must be pointer")
		return
	}
	if vB.Kind() == reflect.Ptr {
		vB = vB.Elem()
	} else {
		err = fmt.Errorf("must be pointer")
		return
	}
	if vA.Kind() != reflect.Struct || vB.Kind() != reflect.Struct {
		err = fmt.Errorf("must be struct")
		return
	}
	tA := vA.Type()
	for i := 0; i < vA.NumField(); i++ {
		fA := tA.Field(i)
		if fA.Name == "ID" {
			continue
		}
		fB := vB.FieldByName(fA.Name)
		if !fB.IsValid() || !fB.CanSet() {
			continue
		}
		fvA := vA.Field(i)
		if fvA.Type().AssignableTo(fB.Type()) {
			fB.Set(fvA)
		}
	}
	return
}

type Pair struct {
	mA any
	mB any
}
