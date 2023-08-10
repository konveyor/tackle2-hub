package migrationwave

import (
	"strconv"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestMigrationWaveCRUD(t *testing.T) {
	for _, r := range Samples {
		for _, app := range r.Applications {
			expectedApp := api.Application{
				Name:        app.Name,
				Description: "Sample application",
			}
			assert.Must(t, Application.Create(&expectedApp))
		}

		for _, stakeholder := range r.Stakeholders {
			expectedStakeholder := api.Stakeholder{
				Name:  stakeholder.Name,
				Email: "sample@example.com",
			}
			assert.Must(t, Stakeholder.Create(&expectedStakeholder))
		}

		for _, stakeholderGroup := range r.StakeholderGroups {
			expectedStakeholderGroup := api.StakeholderGroup{
				Name:        stakeholderGroup.Name,
				Description: "Sample Stakeholder Group",
			}
			assert.Must(t, StakeholderGroup.Create(&expectedStakeholderGroup))
		}

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
		for _, app := range r.Applications {
			assert.Must(t, Application.Delete(app.ID))
		}
		for _, stakeholder := range r.Stakeholders {
			assert.Must(t, Stakeholder.Delete(stakeholder.ID))
		}
		for _, stakeholderGroup := range r.StakeholderGroups {
			assert.Must(t, StakeholderGroup.Delete(stakeholderGroup.ID))
		}
		assert.Must(t, MigrationWave.Delete(r.ID))

		// Check if the MigrationWave is present even after deletion or not.
		_, err = MigrationWave.Get(r.ID)
		if err == nil {
			t.Errorf("Resource exits, but should be deleted: %v", r)
		}
	}
}

func TestMigrationWaveList(t *testing.T) {
	createdMigrationWaves := []api.MigrationWave{}

	for _, r := range Samples {
		for _, app := range r.Applications {
			expectedApp := api.Application{
				Name:        app.Name,
				Description: "Sample application",
			}
			assert.Must(t, Application.Create(&expectedApp))
		}

		for _, stakeholder := range r.Stakeholders {
			expectedStakeholder := api.Stakeholder{
				Name:  stakeholder.Name,
				Email: "sample1@example.com",
			}
			assert.Must(t, Stakeholder.Create(&expectedStakeholder))
		}

		for i, stakeholderGroup := range r.StakeholderGroups {
			expectedStakeholderGroup := api.StakeholderGroup{
				Name:        stakeholderGroup.Name + strconv.Itoa(i),
				Description: "Sample Stakeholder Group",
			}
			assert.Must(t, StakeholderGroup.Create(&expectedStakeholderGroup))
		}
		assert.Must(t, MigrationWave.Create(&r))
		createdMigrationWaves = append(createdMigrationWaves, r)
	}

	// List MigrationWaves.
	got, err := MigrationWave.List()
	if err != nil {
		t.Errorf(err.Error())
	}

	// Compare contents of migration waves.
	for _, createdWave := range createdMigrationWaves {
		found := false
		for _, retrievedWave := range got {
			if assert.FlatEqual(createdWave.ID, retrievedWave.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected Wave not found in the list: %v", createdWave)
		}
	}

	// Delete created resources.
	for _, createdMigrationWave := range createdMigrationWaves {
		for _, app := range createdMigrationWave.Applications {
			assert.Must(t, Application.Delete(app.ID))
		}

		for _, stakeholder := range createdMigrationWave.Stakeholders {
			assert.Must(t, Stakeholder.Delete(stakeholder.ID))
		}

		for _, stakeholderGroup := range createdMigrationWave.StakeholderGroups {
			assert.Must(t, StakeholderGroup.Delete(stakeholderGroup.ID))
		}
		assert.Must(t, MigrationWave.Delete(createdMigrationWave.ID))
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
