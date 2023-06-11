package dependency

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestDependencyCRUD(t *testing.T) {
	for _, sample := range Samples {
		t.Run(fmt.Sprintf("Dependency from %s -> %s", sample.ApplicationFrom.Name, sample.ApplicationTo.Name), func(t *testing.T) {

			assert.Must(t, Application.Create(&sample.ApplicationFrom))
			assert.Must(t, Application.Create(&sample.ApplicationTo))

			// Create a check for reverse dependency

			// Create.
			dependency := api.Dependency{
				From: api.Ref{
					ID: sample.ApplicationFrom.ID,
				},
				To: api.Ref{
					ID: sample.ApplicationTo.ID,
				},
			}
			assert.Should(t, Dependency.Create(&dependency))

			// Get.
			got, err := Dependency.Get(dependency.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !assert.FlatEqual(got.ID, dependency.ID) {
				t.Errorf("Different response error. Got %v, expected %v", got, dependency.ID)
			}

			// Delete dependency.
			assert.Should(t, Dependency.Delete(dependency.ID))

			_, err = Dependency.Get(dependency.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", dependency)
			}

			//Delete Applications
			assert.Must(t, Application.Delete(sample.ApplicationFrom.ID))
			assert.Must(t, Application.Delete(sample.ApplicationTo.ID))
		})
	}
}

func TestDependencyList(t *testing.T) {

	// an array of created dependencies to track them later
	createdDependencies := []api.Dependency{}

	for _, r := range Samples {

		// Create applications.
		assert.Must(t, Application.Create(&r.ApplicationFrom))
		assert.Must(t, Application.Create(&r.ApplicationTo))

		// Create dependencies.
		dependency := api.Dependency{
			From: api.Ref{
				ID: r.ApplicationFrom.ID,
			},
			To: api.Ref{
				ID: r.ApplicationTo.ID,
			},
		}
		assert.Should(t, Dependency.Create(&dependency))
		createdDependencies = append(createdDependencies, dependency)
	}

	// List dependencies.
	got, err := Dependency.List()
	if err != nil {
		t.Errorf(err.Error())
	}

	// check if created Dependencies are in the list we got from Dependency.List()
	for _, createdDependency := range createdDependencies {
		found := false
		for _, retrievedDependency := range got {
			if assert.FlatEqual(createdDependency.ID, retrievedDependency.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected dependency not found in the list: %v", createdDependency)
		}
	}

	// Delete Dependencies and Applications.
	for _, dependency := range createdDependencies {
		assert.Should(t, Dependency.Delete(dependency.ID))
		assert.Must(t, Application.Delete(dependency.From.ID))
		assert.Must(t, Application.Delete(dependency.To.ID))
	}
}

func TestReverseDependency(t *testing.T) {
	for _, r := range ReverseSamples {
		// Create Applications.
		assert.Must(t, Application.Create(&r.Application1))
		assert.Must(t, Application.Create(&r.Application2))
		assert.Must(t, Application.Create(&r.Application3))

		firstDependency := api.Dependency{
			From: api.Ref{
				ID: r.Application1.ID,
			},
			To: api.Ref{
				ID: r.Application2.ID,
			},
		}
		err := Dependency.Create(&firstDependency)
		if err != nil {
			t.Errorf(err.Error())
		}

		secondDependency := api.Dependency{
			From: api.Ref{
				ID: r.Application2.ID,
			},
			To: api.Ref{
				ID: r.Application3.ID,
			},
		}
		err = Dependency.Create(&secondDependency)
		if err != nil {
			t.Errorf(err.Error())
		}

		// Indirect Reverse dependency should fail.
		thirdDependency := api.Dependency{
			From: api.Ref{
				ID: r.Application3.ID,
			},
			To: api.Ref{
				ID: r.Application1.ID,
			},
		}
		err = Dependency.Create(&thirdDependency)
		if err == nil {
			t.Error("Indirect Reverse dependency not allowed")
		}

		fourthDependency := api.Dependency{
			From: api.Ref{
				ID: r.Application1.ID,
			},
			To: api.Ref{
				ID: r.Application3.ID,
			},
		}
		err = Dependency.Create(&fourthDependency)
		if err != nil {
			t.Errorf(err.Error())
		}

		// Direct Reverse dependency should fail.
		fifthDependency := api.Dependency{
			From: api.Ref{
				ID: r.Application2.ID,
			},
			To: api.Ref{
				ID: r.Application1.ID,
			},
		}
		err = Dependency.Create(&fifthDependency)
		if err == nil {
			t.Error("Direct Reverse dependency not allowed")
		}

		// Direct Reverse dependency should fail.
		sixthDependency := api.Dependency{
			From: api.Ref{
				ID: r.Application3.ID,
			},
			To: api.Ref{
				ID: r.Application2.ID,
			},
		}
		err = Dependency.Create(&sixthDependency)
		if err == nil {
			t.Error("Direct Reverse dependency not allowed")
		}

		// Delete Dependencies.
		assert.Should(t, Dependency.Delete(firstDependency.ID))
		assert.Should(t, Dependency.Delete(secondDependency.ID))
		assert.Should(t, Dependency.Delete(thirdDependency.ID))
		assert.Should(t, Dependency.Delete(fourthDependency.ID))
		assert.Should(t, Dependency.Delete(fifthDependency.ID))
		assert.Should(t, Dependency.Delete(sixthDependency.ID))

		// Delete Applications.
		assert.Must(t, Application.Delete(r.Application1.ID))
		assert.Must(t, Application.Delete(r.Application2.ID))
		assert.Must(t, Application.Delete(r.Application3.ID))
	}
}
