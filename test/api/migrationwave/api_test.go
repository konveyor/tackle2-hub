package migrationwave

import (
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestMigrationWaveCRUD(t *testing.T) {
	for _, r := range Samples {
		createdApps := []api.Application{}
		for _, app := range r.Applications {
			expectedApp := api.Application{
				Name:        app.Name,
				Description: "Sample application",
			}
			assert.Must(t, Application.Create(&expectedApp))
			createdApps = append(createdApps, expectedApp)
			r.Applications[0].ID = expectedApp.ID
		}

		createdStakeholders := []api.Stakeholder{}
		for _, stakeholder := range r.Stakeholders {
			expectedStakeholder := api.Stakeholder{
				Name:  stakeholder.Name,
				Email: "sample@example.com",
			}
			assert.Must(t, Stakeholder.Create(&expectedStakeholder))
			createdStakeholders = append(createdStakeholders, expectedStakeholder)
			r.Stakeholders[0].ID = expectedStakeholder.ID
		}

		createdStakeholderGroups := []api.StakeholderGroup{}
		for _, stakeholderGroup := range r.StakeholderGroups {
			expectedStakeholderGroup := api.StakeholderGroup{
				Name:        stakeholderGroup.Name,
				Description: "Sample Stakeholder Group",
			}
			assert.Must(t, StakeholderGroup.Create(&expectedStakeholderGroup))
			createdStakeholderGroups = append(createdStakeholderGroups, expectedStakeholderGroup)
			r.StakeholderGroups[0].ID = expectedStakeholderGroup.ID
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
		assert.Must(t, MigrationWave.Delete(r.ID))

		for _, stakeholderGroup := range createdStakeholderGroups {
			assert.Must(t, StakeholderGroup.Delete(stakeholderGroup.ID))
		}

		for _, stakeholder := range createdStakeholders {
			assert.Must(t, Stakeholder.Delete(stakeholder.ID))
		}

		for _, app := range createdApps {
			assert.Must(t, Application.Delete(app.ID))
		}

		// Check if the MigrationWave is present even after deletion or not.
		_, err = MigrationWave.Get(r.ID)
		if err == nil {
			t.Errorf("Resource exits, but should be deleted: %v", r)
		}
	}
}

func TestMigrationWaveList(t *testing.T) {

	createdMigrationWaves := []api.MigrationWave{}
	createdApps := []api.Application{}
	createdStakeholders := []api.Stakeholder{}
	createdStakeholderGroups := []api.StakeholderGroup{}

	for _, r := range Samples {
		for _, app := range r.Applications {
			expectedApp := api.Application{
				Name:        app.Name,
				Description: "Sample application",
			}
			assert.Must(t, Application.Create(&expectedApp))
			createdApps = append(createdApps, expectedApp)
			r.Applications[0].ID = expectedApp.ID
		}

		for _, stakeholder := range r.Stakeholders {
			expectedStakeholder := api.Stakeholder{
				Name:  stakeholder.Name,
				Email: "sample@example.com",
			}
			assert.Must(t, Stakeholder.Create(&expectedStakeholder))
			createdStakeholders = append(createdStakeholders, expectedStakeholder)
			r.Stakeholders[0].ID = expectedStakeholder.ID
		}

		for _, stakeholderGroup := range r.StakeholderGroups {
			expectedStakeholderGroup := api.StakeholderGroup{
				Name:        stakeholderGroup.Name,
				Description: "Sample Stakeholder Group",
			}
			assert.Must(t, StakeholderGroup.Create(&expectedStakeholderGroup))
			createdStakeholderGroups = append(createdStakeholderGroups, expectedStakeholderGroup)
			r.StakeholderGroups[0].ID = expectedStakeholderGroup.ID
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

	for _, r := range createdMigrationWaves {
		assert.Must(t, MigrationWave.Delete(r.ID))
	}

	for _, stakeholderGroup := range createdStakeholderGroups {
		assert.Must(t, StakeholderGroup.Delete(stakeholderGroup.ID))
	}

	for _, stakeholder := range createdStakeholders {
		assert.Must(t, Stakeholder.Delete(stakeholder.ID))
	}

	for _, app := range createdApps {
		assert.Must(t, Application.Delete(app.ID))
	}
}

func AssertEqualMigrationWaves(t *testing.T, got *api.MigrationWave, expected api.MigrationWave) {
	if got.Name != expected.Name {
		t.Errorf("Different MigrationWave Name Got %v, expected %v", got.Name, expected.Name)
	}
	if got.StartDate != expected.StartDate {
		t.Errorf("Different Start Date Got %v, expected %v", got.StartDate, expected.StartDate)
	}
	for _, expectedApp := range expected.Applications {
		found := false
		for _, gotApp := range got.Applications {
			if assert.FlatEqual(expectedApp.ID, gotApp.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected Wave not found in the list: %v", expectedApp)
		}
	}
	for _, expectedStakeholders := range expected.Stakeholders {
		found := false
		for _, gotStakeholders := range got.Stakeholders {
			if assert.FlatEqual(expectedStakeholders.ID, gotStakeholders.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected Wave not found in the list: %v", expectedStakeholders)
		}
	}
	for _, expectedStakeholderGroup := range expected.StakeholderGroups {
		found := false
		for _, gotStakeholderGroup := range got.StakeholderGroups {
			if assert.FlatEqual(expectedStakeholderGroup.ID, gotStakeholderGroup.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected Wave not found in the list: %v", expectedStakeholderGroup)
		}
	}
}
