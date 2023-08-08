package migrationwave

import (
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestReviewCRUD(t *testing.T) {
	for _, r := range Samples {
		expectedApp := api.Application{
			Name:        "Sample Application",
			Description: "Sample application",
		}
		assert.Must(t, Application.Create(&expectedApp))

		expectedStakeholder := api.Stakeholder{
			Name:  "Sample Stakeholder",
			Email: "sample@example.com",
		}
		assert.Must(t, Stakeholder.Create(&expectedStakeholder))

		expectedStakeholderGroup := api.StakeholderGroup{
			Name: "Sample Stakeholder Group",
		}
		assert.Must(t, StakeholderGroup.Create(&expectedStakeholderGroup))

		assert.Must(t, MigrationWave.Create(&r))

		// Get migration wave.
		got, err := MigrationWave.Get(r.ID)
		if err != nil {
			t.Errorf(err.Error())
		}

		// Compare got values with expected values.
		AssertEqualMigrationWaves(t, got, r)

		// Update MigrationWave's Name.
		r.EndDate = r.EndDate.Add(30 * time.Minute)
		assert.Should(t, MigrationWave.Update(&r))

		// Find MigrationWave and check its parameters with the got(On Updation).
		got, err = MigrationWave.Get(r.ID)
		if err != nil {
			t.Errorf(err.Error())
		}

		// Check if the unchanged values remain same or not.
		AssertEqualMigrationWaves(t, got, r)

		// Delete created Applications, Stakeholders,StakeholdersGroup and MigrationWave
		assert.Must(t, Application.Delete(expectedApp.ID))
		assert.Must(t, Stakeholder.Delete(expectedStakeholder.ID))
		assert.Must(t, StakeholderGroup.Delete(expectedStakeholderGroup.ID))
		assert.Must(t, MigrationWave.Delete(r.ID))

		// Check if the MigrationWave is present even after deletion or not.
		_, err = MigrationWave.Get(r.ID)
		if err == nil {
			t.Errorf("Resource exits, but should be deleted: %v", r)
		}
	}
}

func AssertEqualMigrationWaves(t *testing.T, got *api.MigrationWave, expected api.MigrationWave) {
	if got.Name != expected.Name {
		t.Errorf("Different MigrationWave Name Got %v, expected %v", got.Name, expected.Name)
	}
	if got.StartDate != expected.StartDate {
		t.Errorf("Different Start Date Got %v, expected %v", got.StartDate, expected.StartDate)
	}
}
