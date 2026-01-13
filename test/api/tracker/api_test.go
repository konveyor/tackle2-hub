package tracker

import (
	"strconv"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTrackerCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("Tracker "+r.Kind+" CRUD", func(t *testing.T) {
			// Create a sample identity for the tracker.
			identity := api.Identity{
				Kind: r.Kind,
				Name: r.Identity.Name,
			}
			assert.Must(t, Identity.Create(&identity))
			// Copy the identity name to the tracker.
			r.Identity.ID = identity.ID

			// Create a tracker.
			assert.Must(t, Tracker.Create(&r))

			// Get.
			got, err := Tracker.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare got values with expected values.
			AssertEqualTrackers(t, got, r)

			// Update Message.
			r.Message = "Updated Comment " + r.Message
			assert.Should(t, Tracker.Update(&r))

			// Find Tracker and check its parameters with the got(On Updation).
			got, err = Tracker.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check if the unchanged values remain same or not.
			AssertEqualTrackers(t, got, r)

			// Delete identity and tracker.
			assert.Must(t, Tracker.Delete(r.ID))
			assert.Must(t, Identity.Delete(identity.ID))

			// Check if the Tracker is present even after deletion or not.
			_, err = Tracker.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})

		t.Run("Tracker "+r.Kind+" Project", func(t *testing.T) {
			// Create a sample identity for the tracker.
			identity := api.Identity{
				Kind: r.Kind,
				Name: r.Identity.Name,
			}
			assert.Must(t, Identity.Create(&identity))

			// Copy the identity name to the tracker.
			r.Identity.ID = identity.ID

			// Create a tracker.
			assert.Must(t, Tracker.Create(&r))

			projectsList, err := Tracker.ListProjects(r.ID)
			if err != nil {
				// check for type of service (maybe Connected?)
				// if API service then print Not connected to Jira and pass the test
				// else fail the test
				if r.Connected == false {
					t.Logf("Not connected to Jira(Thus passing the API Test)")
				} else {
					t.Errorf(err.Error())
				}
			}

			for _, projects := range projectsList {

				// convert project Id's to uint.
				projectID, err := strconv.Atoi(projects.ID)
				if err != nil {
					t.Errorf(err.Error())
				}

				_, err = Tracker.GetProjects(r.ID, uint(projectID))
				if err != nil {
					if r.Connected == false {
						t.Logf("Not connected to Jira(Thus passing the API Test)")
					} else {
						t.Errorf(err.Error())
					}
				}

				_, err = Tracker.ListProjectIssueTypes(r.ID, uint(projectID))
				if err != nil {
					if r.Connected == false {
						t.Logf("Not connected to Jira(Thus passing the API Test)")
					} else {
						t.Errorf(err.Error())
					}
				}
			}

			// Delete identity and tracker.
			assert.Must(t, Tracker.Delete(r.ID))
			assert.Must(t, Identity.Delete(identity.ID))

			// Check if the Tracker is present even after deletion or not.
			_, err = Tracker.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestTrackerList(t *testing.T) {
	createdTrackers := []api.Tracker{}

	for _, r := range Samples {
		identity := api.Identity{
			Kind: r.Kind,
			Name: r.Identity.Name,
		}
		assert.Must(t, Identity.Create(&identity))
		r.Identity.ID = identity.ID

		assert.Must(t, Tracker.Create(&r))
		createdTrackers = append(createdTrackers, r)
	}

	// List Trackers.
	got, err := Tracker.List()
	if err != nil {
		t.Errorf(err.Error())
	}

	// check if created Trackers are in the list we got from Tracker.List()
	for _, createdTracker := range createdTrackers {
		found := false
		for _, retrievedTracker := range got {
			if assert.FlatEqual(createdTracker.ID, retrievedTracker.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tracker not found in the list: %v", createdTracker)
		}
	}

	// Delete related trackers and identities.
	for _, tracker := range createdTrackers {
		assert.Must(t, Tracker.Delete(tracker.ID))
		assert.Must(t, Identity.Delete(tracker.Identity.ID))
	}
}

func AssertEqualTrackers(t *testing.T, got *api.Tracker, expected api.Tracker) {
	if got.Name != expected.Name {
		t.Errorf("Different Tracker Name Got %v, expected %v", got.Name, expected.Name)
	}
	if got.URL != expected.URL {
		t.Errorf("Different Url Got %v, expected %v", got.URL, expected.URL)
	}
	if got.Kind != expected.Kind {
		t.Errorf("Different Kind Got %v, expected %v", got.Kind, expected.Kind)
	}
	if got.Connected != expected.Connected {
		t.Errorf("Different Connected Got %v, expected %v", got.Connected, expected.Connected)
	}
	if got.Identity.Name != expected.Identity.Name {
		t.Errorf("Different Identity's Name Got %v, expected %v", got.Identity.Name, expected.Identity.Name)
	}
	if got.Insecure != expected.Insecure {
		t.Errorf("Different Insecure Got %v, expected %v", got.Insecure, expected.Kind)
	}
}
