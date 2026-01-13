package seed

import (
	"errors"
	"io/fs"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/shared/settings"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
)

var (
	Settings = &settings.Settings
	log      = logr.New("seeding", Settings.Log.Migration)
)

const (
	// SeedKey identifies the setting containing the applied seed digest.
	SeedKey = ".hub.db.seed"
	// BuildKey identifies setting for the hub build that seeded.
	BuildKey = SeedKey + ".build"
)

// Seeder specifies an interface for seeding DB models.
type Seeder interface {
	With(libseed.Seed) error
	Apply(*gorm.DB) error
}

// Seed applies DB seeds.
func Seed() (err error) {
	var db *gorm.DB

	db, err = database.Open(true)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = database.Close(db)
		}
	}()

	seeds, checksum, err := libseed.ReadFromDir(settings.Settings.Hub.DB.SeedPath, libseed.AllVersions)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Info("Seed directory not found.")
			err = nil
			return
		}
		err = liberr.Wrap(err)
		return
	}
	if len(seeds) == 0 {
		log.Info("No seed files found.")
		return
	}

	skipped, err := skip(db, checksum)
	if err != nil {
		return
	}
	if skipped {
		log.Info("Seeding skipped.")
		return
	}

	log.Info("Applying seeds.")
	seeder := Hub{}
	for _, seed := range seeds {
		err = seeder.With(seed)
		if err != nil {
			return
		}
	}

	err = db.Transaction(func(tx *gorm.DB) (err error) {
		err = seeder.Apply(tx)
		if err != nil {
			return
		}
		err = saveChecksum(tx, checksum)
		if err != nil {
			return
		}
		err = saveBuild(tx)
		if err != nil {
			return
		}
		return
	})
	return
}

// skip returns true when seeding can be skipped.
func skip(db *gorm.DB, checksum []byte) (skip bool, err error) {
	match, err := compareChecksum(db, checksum)
	if err != nil {
		return
	}
	if !match {
		return
	}
	match, err = matchBuild(db)
	if err != nil {
		return
	}
	if !match {
		return
	}
	skip = true
	return
}
