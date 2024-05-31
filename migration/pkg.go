package migration

import (
	"github.com/jortel/go-utils/logr"
	v10 "github.com/konveyor/tackle2-hub/migration/v10"
	v11 "github.com/konveyor/tackle2-hub/migration/v11"
	v12 "github.com/konveyor/tackle2-hub/migration/v12"
	v13 "github.com/konveyor/tackle2-hub/migration/v13"
	v2 "github.com/konveyor/tackle2-hub/migration/v2"
	v3 "github.com/konveyor/tackle2-hub/migration/v3"
	v4 "github.com/konveyor/tackle2-hub/migration/v4"
	v5 "github.com/konveyor/tackle2-hub/migration/v5"
	v6 "github.com/konveyor/tackle2-hub/migration/v6"
	v7 "github.com/konveyor/tackle2-hub/migration/v7"
	v8 "github.com/konveyor/tackle2-hub/migration/v8"
	v9 "github.com/konveyor/tackle2-hub/migration/v9"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
)

var log = logr.WithName("migration")
var Settings = &settings.Settings

// VersionKey is the setting containing the migration version.
const VersionKey = ".migration.version"

// MinimumVersion is the index of the
// earliest version that we can migrate from.
var MinimumVersion = 1

// Version represents the value of the .migration.version setting.
type Version struct {
	Version int `json:"version"`
}

// Migration encapsulates the functionality necessary to perform a migration.
type Migration interface {
	Apply(*gorm.DB) error
	Models() []interface{}
}

// All migrations in order.
func All() []Migration {
	return []Migration{
		v2.Migration{},
		v3.Migration{},
		v4.Migration{},
		v5.Migration{},
		v6.Migration{},
		v7.Migration{},
		v8.Migration{},
		v9.Migration{},
		v10.Migration{},
		v11.Migration{},
		v12.Migration{},
		v13.Migration{},
	}
}
