package migration

import (
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/migration/v2"
	"github.com/konveyor/tackle2-hub/migration/v3"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
)

var log = logging.WithName("migration")
var Settings = &settings.Settings

//
// VersionKey is the setting containing the migration version.
const VersionKey = ".migration.version"

//
// MinimumVersion is the index of the
// earliest version that we can migrate from.
var MinimumVersion = 1

//
// Version represents the value of the .migration.version setting.
type Version struct {
	Version int `json:"version"`
}

//
// Migration encapsulates the functionality necessary to perform a migration.
type Migration interface {
	Apply(*gorm.DB) error
	Models() []interface{}
}

//
// All migrations in order.
func All() []Migration {
	return []Migration{
		v2.Migration{},
		v3.Migration{},
	}
}
