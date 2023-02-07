package migration

import (
	"encoding/json"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/onsi/gomega"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestFreshInstall(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	Settings.DB.Path = "/tmp/freshinstall.db"
	_ = os.Remove(Settings.DB.Path)

	MinimumVersion = 2
	migrations := []Migration{
		&TestMigration{Version: 3, ShouldRun: true},
		&TestMigration{Version: 4, ShouldRun: true},
		&TestMigration{Version: 5, ShouldRun: true},
	}


	err := Migrate(migrations)
	g.Expect(err).To(gomega.BeNil())

	for _, m := range migrations {
		migration := m.(*TestMigration)
		g.Expect(migration.Ran).To(gomega.Equal(migration.ShouldRun))
	}
	expectVersion(g, 5)

	_ = os.Remove(Settings.DB.Path)
}

func TestUpgrade(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	Settings.DB.Path = "/tmp/existinginstall.db"
	_ = os.Remove(Settings.DB.Path)
	setup(g, 3)

	MinimumVersion = 2
	migrations := []Migration{
		&TestMigration{Version: 3, ShouldRun: false},
		&TestMigration{Version: 4, ShouldRun: true},
		&TestMigration{Version: 5, ShouldRun: true},
	}

	err := Migrate(migrations)
	g.Expect(err).To(gomega.BeNil())
	for _, m := range migrations {
		migration := m.(*TestMigration)
		g.Expect(migration.Ran).To(gomega.Equal(migration.ShouldRun))
	}
	expectVersion(g, 5)

	migrations = []Migration{
		&TestMigration{Version: 3, ShouldRun: false},
		&TestMigration{Version: 4, ShouldRun: false},
		&TestMigration{Version: 5, ShouldRun: false},
	}
	err = Migrate(migrations)
	g.Expect(err).To(gomega.BeNil())
	for _, m := range migrations {
		migration := m.(*TestMigration)
		g.Expect(migration.Ran).To(gomega.Equal(migration.ShouldRun))
	}
	expectVersion(g, 5)

	_ = os.Remove(Settings.DB.Path)
}

func TestUnsupportedVersion(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	Settings.DB.Path = "/tmp/unsupported.db"
	_ = os.Remove(Settings.DB.Path)
	setup(g, 1)

	MinimumVersion = 2
	migrations := []Migration{
		&TestMigration{Version: 3, ShouldRun: false},
		&TestMigration{Version: 4, ShouldRun: false},
		&TestMigration{Version: 5, ShouldRun: false},
	}
	err := Migrate(migrations)
	g.Expect(err.Error()).To(gomega.Equal("unsupported database version"))
	for _, m := range migrations {
		migration := m.(*TestMigration)
		g.Expect(migration.Ran).To(gomega.Equal(migration.ShouldRun))
	}

	expectVersion(g, 1)

	_ = os.Remove(Settings.DB.Path)
}

type TestMigration struct {
	Version   int
	ShouldRun bool
	Ran       bool
}

func (r *TestMigration) Apply(db *gorm.DB) (err error) {
	r.Ran = true
	return
}

func (r *TestMigration) Models() (models []interface{}) {
	return
}

func setup(g *gomega.GomegaWithT, version int) {
	db, err := database.Open(false)
	g.Expect(err).To(gomega.BeNil())
	result := db.Create(&model.Setting{Key: VersionKey})
	g.Expect(result.Error).To(gomega.BeNil())
	err = setVersion(db, version)
	g.Expect(err).To(gomega.BeNil())
	err = database.Close(db)
	g.Expect(err).To(gomega.BeNil())
}

func expectVersion(g *gomega.GomegaWithT, version int) {
	db, err := database.Open(false)
	g.Expect(err).To(gomega.BeNil())
	setting := &model.Setting{}
	result := db.Find(setting, "key", VersionKey)
	g.Expect(result.Error).To(gomega.BeNil())
	var v Version
	_ = json.Unmarshal(setting.Value, &v)
	g.Expect(v.Version).To(gomega.Equal(version))
	_ = database.Close(db)
}
