package binding

import (
	"errors"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestMigrationWave(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the migration wave to create
	now := time.Now()
	migrationWave := &api.MigrationWave{
		Name:      "Test Migration Wave",
		StartDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Add(30 * 24 * time.Hour),
	}

	// CREATE: Create the migration wave
	err := client.MigrationWave.Create(migrationWave)
	g.Expect(err).To(BeNil())
	g.Expect(migrationWave.ID).NotTo(BeZero())

	defer func() {
		_ = client.MigrationWave.Delete(migrationWave.ID)
	}()

	// GET: Retrieve the migration wave and verify it matches
	retrieved, err := client.MigrationWave.Get(migrationWave.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(migrationWave, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the migration wave
	migrationWave.Name = "Updated Migration Wave"
	migrationWave.EndDate = migrationWave.EndDate.Add(15 * 24 * time.Hour)

	err = client.MigrationWave.Update(migrationWave)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.MigrationWave.Get(migrationWave.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(migrationWave, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the migration wave
	err = client.MigrationWave.Delete(migrationWave.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.MigrationWave.Get(migrationWave.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
