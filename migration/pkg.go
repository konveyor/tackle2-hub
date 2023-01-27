package migration

import (
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/migration/v1"
	"github.com/konveyor/tackle2-hub/migration/v2"
	"github.com/konveyor/tackle2-hub/migration/v3"
	"gorm.io/gorm"
)

var log = logging.WithName("migration")

//
// VersionKey is the setting containing the migration version.
const VersionKey = ".migration.version"

//
// Version represents the value of the .migration.version setting.
type Version struct {
	Version int `json:"version"`
}

//
// Migration encapsulates the functionality necessary to perform a migration.
type Migration interface {
	Apply(*gorm.DB) error
	Name() string
}

//
// All migrations in order.
func All() []Migration {
	return []Migration{
		v1.Migration{},
		v2.Migration{},
		v3.Migration{},
	}
}
