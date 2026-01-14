package migration

import (
	"github.com/jortel/go-utils/logr"
	v10 "github.com/konveyor/tackle2-hub/internal/migration/v10"
	v11 "github.com/konveyor/tackle2-hub/internal/migration/v11"
	v12 "github.com/konveyor/tackle2-hub/internal/migration/v12"
	v13 "github.com/konveyor/tackle2-hub/internal/migration/v13"
	v14 "github.com/konveyor/tackle2-hub/internal/migration/v14"
	v15 "github.com/konveyor/tackle2-hub/internal/migration/v15"
	v16 "github.com/konveyor/tackle2-hub/internal/migration/v16"
	v17 "github.com/konveyor/tackle2-hub/internal/migration/v17"
	v18 "github.com/konveyor/tackle2-hub/internal/migration/v18"
	v19 "github.com/konveyor/tackle2-hub/internal/migration/v19"
	v2 "github.com/konveyor/tackle2-hub/internal/migration/v2"
	v20 "github.com/konveyor/tackle2-hub/internal/migration/v20"
	v21 "github.com/konveyor/tackle2-hub/internal/migration/v21"
	v3 "github.com/konveyor/tackle2-hub/internal/migration/v3"
	v4 "github.com/konveyor/tackle2-hub/internal/migration/v4"
	v5 "github.com/konveyor/tackle2-hub/internal/migration/v5"
	v6 "github.com/konveyor/tackle2-hub/internal/migration/v6"
	v7 "github.com/konveyor/tackle2-hub/internal/migration/v7"
	v8 "github.com/konveyor/tackle2-hub/internal/migration/v8"
	v9 "github.com/konveyor/tackle2-hub/internal/migration/v9"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
)

var Log = logr.New("migration", Settings.Log.Migration)
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
	Models() []any
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
		v14.Migration{},
		v15.Migration{},
		v16.Migration{},
		v17.Migration{},
		v18.Migration{},
		v19.Migration{},
		v20.Migration{},
		v21.Migration{},
	}
}
