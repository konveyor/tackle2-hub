package seed

import (
	"fmt"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/migration"
	"github.com/konveyor/tackle2-hub/internal/model"
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
	Generator
}

// With collects the resources to be seeded.
func (r *Hub) With(seed libseed.Seed) (err error) {
	kind := strings.ToLower(seed.Kind)
	switch kind {
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
	case libseed.KindGenerator:
		err = r.Generator.With(seed)
	default:
		log.Info("WARNING: " + kind + " not supported.")
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
	err = r.Generator.Apply(db)
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
	log.Info("Seed checksum", "matched", match)
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

// matchBuild
func matchBuild(db *gorm.DB) (matched bool, err error) {
	build, err := getBuild(db)
	if err != nil {
		return
	}
	if build == "" {
		return
	}
	matched = build == Settings.Hub.Build
	log.Info("Seed build (version)", "matched", matched)
	return
}

// getBuild returns the hub build version that seeded.
func getBuild(db *gorm.DB) (version string, err error) {
	setting := &model.Setting{
		Key: BuildKey,
	}
	db = db.Where("key", BuildKey)
	err = db.FirstOrCreate(setting).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if n, cast := setting.Value.(string); cast {
		version = n
	}
	return
}

// saveBuild update settings with current build that seeded.
func saveBuild(db *gorm.DB) (err error) {
	setting := &model.Setting{
		Key:   BuildKey,
		Value: Settings.Hub.Build,
	}
	db = db.Where("key", BuildKey)
	err = db.Updates(setting).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
