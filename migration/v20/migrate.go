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

var Models []Pair = []Pair{
	// Models with no dependencies
	{mA: v19.TagCategory{}, mB: model.TagCategory{}},
	{mA: v19.Identity{}, mB: model.Identity{}, renamed: Map{"Default": "IsDefault", "User": "Userid"}},
	{mA: v19.JobFunction{}, mB: model.JobFunction{}},
	{mA: v19.StakeholderGroup{}, mB: model.StakeholderGroup{}},
	{mA: v19.File{}, mB: model.File{}},
	{mA: v19.Setting{}, mB: model.Setting{}},
	{mA: v19.Bucket{}, mB: model.Bucket{}},
	// Models with dependencies
	{mA: v19.Stakeholder{}, mB: model.Stakeholder{}},                                              // depends on JobFunction
	{mA: v19.MigrationWave{}, mB: model.MigrationWave{}},                                          // depends on Stakeholders, Groups
	{mA: v19.BusinessService{}, mB: model.BusinessService{}},                                      // depends on Stakeholder
	{mA: v19.Application{}, mB: model.Application{}, renamed: Map{"Binary": "BinaryCoordinates"}}, // depends on BusinessService, Owner, Platform, MigrationWave, Bucket
	{mA: v19.TaskGroup{}, mB: model.TaskGroup{}},                                                  // depends on Bucket
	{mA: v19.Task{}, mB: model.Task{}},                                                            // depends on Application, Platform, TaskGroup, Bucket
	{mA: v19.Manifest{}, mB: model.Manifest{}},                                                    // depends on Application
	{mA: v19.Tracker{}, mB: model.Tracker{}},                                                      // depends on Identity
	{mA: v19.Ticket{}, mB: model.Ticket{}},                                                        // depends on Application, Tracker
	{mA: v19.Analysis{}, mB: model.Analysis{}},                                                    // depends on Application
	{mA: v19.TechDependency{}, mB: model.TechDependency{}},                                        // depends on Analysis
	{mA: v19.Insight{}, mB: model.Insight{}},                                                      // depends on Analysis
	{mA: v19.Incident{}, mB: model.Incident{}},                                                    // depends on Insight
	{mA: v19.RuleSet{}, mB: model.RuleSet{}},                                                      // depends on Identity and itself (many2many)
	{mA: v19.Rule{}, mB: model.Rule{}},                                                            // depends on RuleSet, File
	{mA: v19.Target{}, mB: model.Target{}},                                                        // depends on RuleSet, File
	{mA: v19.Archetype{}, mB: model.Archetype{}},                                                  // depends on Review
	{mA: v19.TargetProfile{}, mB: model.TargetProfile{}},                                          // depends on Archetype
	{mA: v19.Generator{}, mB: model.Generator{}},                                                  // depends on Identity
	{mA: v19.ProfileGenerator{}, mB: model.ProfileGenerator{}},                                    // depends on TargetProfile, Generator
	{mA: v19.Questionnaire{}, mB: model.Questionnaire{}},                                          // depends on nothing internal
	{mA: v19.Assessment{}, mB: model.Assessment{}},                                                // depends on Application, Archetype, Questionnaire
	{mA: v19.ImportSummary{}, mB: model.ImportSummary{}},                                          // depends on nothing internal
	{mA: v19.Import{}, mB: model.Import{}},                                                        // depends on ImportSummary
	{mA: v19.ImportTag{}, mB: model.ImportTag{}},                                                  // depends on Import
	{mA: v19.Fact{}, mB: model.Fact{}},                                                            // depends on Application
	{mA: v19.Tag{}, mB: model.Tag{}},                                                              // depends on TagCategory
	{mA: v19.ApplicationTag{}, mB: model.ApplicationTag{}},                                        // depends on Application, Tag
	{mA: v19.ApplicationIdentity{}, mB: model.ApplicationIdentity{}},                              // depends on Application, Identity
	{mA: v19.Platform{}, mB: model.Platform{}},                                                    // depends on Identity
	{mA: v19.TaskReport{}, mB: model.TaskReport{}},                                                // depends on Task
	{mA: v19.Review{}, mB: model.Review{}},                                                        // depends on Application, Archetype
	{mA: v19.Dependency{}, mB: model.Dependency{}},                                                // depends on Application
	{mA: v19.Proxy{}, mB: model.Proxy{}},                                                          // depends on Identity
}

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
	err = db.Transaction(func(tx *gorm.DB) (err error) {
		// migration v7 deleted seeded ruleSet but because
		// FKs disabled, the related rules did not get deleted.
		err = sqlite.Delete(&v19.Rule{}, "id <= ?", 14).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		for _, p := range Models {
			Log.Info(p.String())
			err = r.migratePair(sqlite, db, p)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
		return
	})
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
		p := p.use(mA, mB, p)
		err = r.migrate(p)
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

func (r Migration) migrate(pair Pair) (err error) {
	vA := reflect.ValueOf(pair.mA)
	vB := reflect.ValueOf(pair.mB)
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

type Map map[string]string

type Pair struct {
	renamed Map
	mA      any
	mB      any
}

func (p Pair) use(mA, mB any, other Pair) Pair {
	return Pair{
		renamed: other.renamed,
		mA:      mA,
		mB:      mB,
	}
}

func (p *Pair) rename(name string) (renamed string) {
	renamed = name
	if p.renamed != nil {
		n, found := p.renamed[name]
		if found {
			renamed = n
		}
	}
	return
}

func (p *Pair) String() (s string) {
	tA := reflect.TypeOf(p.mA)
	tB := reflect.TypeOf(p.mB)
	if tA.Kind() == reflect.Ptr {
		tA = tA.Elem()
	}
	if tB.Kind() == reflect.Ptr {
		tB = tB.Elem()
	}
	s = fmt.Sprintf("%s => %s", tA.Name(), tB.Name())
	return
}
