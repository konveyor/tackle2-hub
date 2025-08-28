package migration

import (
	"errors"
	"fmt"
	"strconv"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	v10 "github.com/konveyor/tackle2-hub/migration/v10"
	v11 "github.com/konveyor/tackle2-hub/migration/v11"
	v12 "github.com/konveyor/tackle2-hub/migration/v12"
	v13 "github.com/konveyor/tackle2-hub/migration/v13"
	v14 "github.com/konveyor/tackle2-hub/migration/v14"
	v15 "github.com/konveyor/tackle2-hub/migration/v15"
	v16 "github.com/konveyor/tackle2-hub/migration/v16"
	v17 "github.com/konveyor/tackle2-hub/migration/v17"
	v18 "github.com/konveyor/tackle2-hub/migration/v18"
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

var Log = logr.WithName("migration")
var Settings = &settings.Settings

// VersionKey is the setting containing the migration version.
const VersionKey = ".migration.version"

// Version represents the value of the .migration.version setting.
type Version struct {
	Version int `json:"version"`
}

func (v *Version) Validate(migrations []Migration) (err error) {
	if v.Version < 1 || v.Version > len(migrations) {
		err = &VersionError{Version: v.Version}
		err = liberr.Wrap(err)
	}
	return
}

func (v *Version) Index() (index int) {
	index = v.Version - 1
	if index < 0 {
		index = 0
	}
	return
}

func (v *Version) With(index int) {
	v.Version = index + 1
}

func (v *Version) String() (s string) {
	s = strconv.Itoa(v.Version)
	return
}

func (v *Version) Latest(migrations []Migration) (latest *Version) {
	latest = &Version{
		Version: len(migrations),
	}
	return
}

func (v *Version) Rewind(n int) *Version {
	v.Version -= n
	if v.Version < 1 {
		v.Version = 1
	}
	return v
}

func (v *Version) Next() *Version {
	v.Version++
	return v
}

// NopMigration placeholder.
type NopMigration struct{}

func (m *NopMigration) Apply(*gorm.DB) (err error) {
	return
}

func (m *NopMigration) Models() (none []any) {
	return
}

type VersionError struct {
	Version int
}

func (v *VersionError) Is(err error) (matched bool) {
	var inst *VersionError
	matched = errors.As(err, &inst)
	return
}

func (v *VersionError) Error() (s string) {
	s = fmt.Sprintf("Migration version=%d not-valid.", v.Version)
	return
}

// Migration encapsulates the functionality necessary to perform a migration.
type Migration interface {
	Apply(*gorm.DB) error
	Models() []any
}

// All migrations in order.
// Note: pruned version MUST not be removed but instead, replaced with NopMigration.
func All() []Migration {
	return []Migration{
		&NopMigration{},
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
	}
}
