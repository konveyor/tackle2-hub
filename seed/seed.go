package seed

import (
	"fmt"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/migration"
	"github.com/konveyor/tackle2-hub/model"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

// Hub is responsible for collecting and applying Hub seeds.
type Hub struct {
	TagCategory
	JobFunction
	RuleSet
	Target
	Questionnaire
}

// With collects the resources to be seeded.
func (r *Hub) With(seed libseed.Seed) (err error) {
	switch strings.ToLower(seed.Kind) {
	case libseed.KindTagCategory:
		err = r.TagCategory.With(seed)
	case libseed.KindJobFunction:
		err = r.JobFunction.With(seed)
	case libseed.KindRuleSet:
		err = r.RuleSet.With(seed)
	case libseed.KindTarget:
		err = r.Target.With(seed)
	case libseed.KindQuestionnaire:
		err = r.Questionnaire.With(seed)
	default:
		err = liberr.New("unknown kind", "kind", seed.Kind, "file", seed.Filename())
	}
	return
}

// Apply seeds the database with resources from the seed files.
func (r *Hub) Apply(db *gorm.DB) (err error) {
	err = r.TagCategory.Apply(db)
	if err != nil {
		return
	}
	err = r.JobFunction.Apply(db)
	if err != nil {
		return
	}
	err = r.RuleSet.Apply(db)
	if err != nil {
		return
	}
	err = r.Target.Apply(db)
	if err != nil {
		return
	}
	err = r.Questionnaire.Apply(db)
	if err != nil {
		return
	}

	return
}

// compareChecksum compares the seed checksum to the stored checksum.
func compareChecksum(db *gorm.DB, checksum []byte) (match bool, err error) {
	setting := &model.Setting{}
	result := db.FirstOrCreate(setting, model.Setting{Key: SeedKey})
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	var seededChecksum string
	err = setting.As(&seededChecksum)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	match = seededChecksum == fmt.Sprintf("%x", checksum)
	return
}

// saveChecksum saves the seed checksum to the setting specified by SeedKey.
func saveChecksum(db *gorm.DB, checksum []byte) (err error) {
	setting := &model.Setting{Key: SeedKey}
	setting.Value = fmt.Sprintf("%x", checksum)
	result := db.Where("key", SeedKey).Updates(setting)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	return
}

// migrationVersion gets the current migration version.
func migrationVersion(db *gorm.DB) (version uint, err error) {
	setting := &model.Setting{}
	result := db.First(setting, model.Setting{Key: migration.VersionKey})
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}

	var v migration.Version
	err = setting.As(&v)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	version = uint(v.Version)
	return
}
