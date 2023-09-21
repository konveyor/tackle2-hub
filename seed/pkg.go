package seed

import (
	"errors"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/settings"
	libseed "github.com/konveyor/tackle2-seed/pkg"
	"gorm.io/gorm"
	"io/fs"
)

var log = logr.WithName("seeding")

//
// SeedKey identifies the setting containing the applied seed digest.
const SeedKey = ".hub.db.seed"

//
// Seeder specifies an interface for seeding DB models.
type Seeder interface {
	With(libseed.Seed) error
	Apply(*gorm.DB) error
}

//
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

	match, err := compareChecksum(db, checksum)
	if err != nil {
		return
	}
	if match {
		log.Info("Seed checksum match.")
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
		return
	})
	return
}
