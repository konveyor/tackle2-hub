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

	// Create application for the migration wave
	application := &api.Application{
		Name:        "Migration Wave App",
		Description: "Application for migration wave testing",
	}
	err := client.Application.Create(application)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

	// Create stakeholder for the migration wave
	stakeholder := &api.Stakeholder{
		Name:  "Wave Stakeholder",
		Email: "stakeholder@wave.local",
	}
	err = client.Stakeholder.Create(stakeholder)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Stakeholder.Delete(stakeholder.ID)
	})

	// Create stakeholder group for the migration wave
	stakeholderGroup := &api.StakeholderGroup{
		Name:        "Wave Group",
		Description: "Stakeholder group for migration wave",
	}
	err = client.StakeholderGroup.Create(stakeholderGroup)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.StakeholderGroup.Delete(stakeholderGroup.ID)
	})

	// Define the migration wave to create
	now := time.Now()
	migrationWave := &api.MigrationWave{
		Name:      "Test Migration Wave",
		StartDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Add(30 * 24 * time.Hour),
		Applications: []api.Ref{
			{ID: application.ID},
		},
		Stakeholders: []api.Ref{
			{ID: stakeholder.ID},
		},
		StakeholderGroups: []api.Ref{
			{ID: stakeholderGroup.ID},
		},
	}

	// CREATE: Create the migration wave
	err = client.MigrationWave.Create(migrationWave)
	g.Expect(err).To(BeNil())
	g.Expect(migrationWave.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.MigrationWave.Delete(migrationWave.ID)
	})

	// GET: List migration waves
	list, err := client.MigrationWave.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(migrationWave, list[0], "Applications.Name", "Stakeholders.Name", "StakeholderGroups.Name")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the migration wave and verify it matches
	retrieved, err := client.MigrationWave.Get(migrationWave.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(migrationWave, retrieved, "Applications.Name", "Stakeholders.Name", "StakeholderGroups.Name")
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
	eq, report = cmp.Eq(migrationWave, updated, "UpdateUser", "Applications.Name", "Stakeholders.Name", "StakeholderGroups.Name")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the migration wave
	err = client.MigrationWave.Delete(migrationWave.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.MigrationWave.Get(migrationWave.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
