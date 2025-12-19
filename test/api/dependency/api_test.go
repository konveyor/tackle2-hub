package dependency

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestDependencyCRUD(t *testing.T) {
	for _, sample := range Samples {
		t.Run(fmt.Sprintf("Dependency from %s -> %s", sample.ApplicationFrom.Name, sample.ApplicationTo.Name), func(t *testing.T) {

			assert.Must(t, Application.Create(&sample.ApplicationFrom))
			assert.Must(t, Application.Create(&sample.ApplicationTo))

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

		firstDependencyPass := api.Dependency{
			From: api.Ref{
				ID: r.Application1.ID,
			},
			To: api.Ref{
				ID: r.Application2.ID,
			},
		}
		assert.Should(t, Dependency.Create(&firstDependencyPass))

		secondDependencyPass := api.Dependency{
			From: api.Ref{
				ID: r.Application2.ID,
			},
			To: api.Ref{
				ID: r.Application3.ID,
			},
		}
		assert.Should(t, Dependency.Create(&secondDependencyPass))

		// Indirect Reverse dependency should fail.
		indirectReverseDependencyFail := api.Dependency{
			From: api.Ref{
				ID: r.Application3.ID,
			},
			To: api.Ref{
				ID: r.Application1.ID,
			},
		}
		err := Dependency.Create(&indirectReverseDependencyFail)
		if err == nil {
			t.Error("Indirect Reverse dependency not allowed")
		}

		thirdDependencyPass := api.Dependency{
			From: api.Ref{
				ID: r.Application1.ID,
			},
			To: api.Ref{
				ID: r.Application3.ID,
			},
		}
		assert.Should(t, Dependency.Create(&thirdDependencyPass))

		// Direct Reverse dependency should fail.
		DirectReverseDependencyFail := api.Dependency{
			From: api.Ref{
				ID: r.Application2.ID,
			},
			To: api.Ref{
				ID: r.Application1.ID,
			},
		}
		err = Dependency.Create(&DirectReverseDependencyFail)
		if err == nil {
			t.Error("Direct Reverse dependency not allowed")
		}

		// Direct Reverse dependency should fail.
		anotherDirectReverseDependencyFail := api.Dependency{
			From: api.Ref{
				ID: r.Application3.ID,
			},
			To: api.Ref{
				ID: r.Application2.ID,
			},
		}
		err = Dependency.Create(&anotherDirectReverseDependencyFail)
		if err == nil {
			t.Error("Direct Reverse dependency not allowed")
		}

		// Delete Dependencies.
		assert.Should(t, Dependency.Delete(firstDependencyPass.ID))
		assert.Should(t, Dependency.Delete(secondDependencyPass.ID))
		assert.Should(t, Dependency.Delete(thirdDependencyPass.ID))

		// Delete Applications.
		assert.Must(t, Application.Delete(r.Application1.ID))
		assert.Must(t, Application.Delete(r.Application2.ID))
		assert.Must(t, Application.Delete(r.Application3.ID))
	}
}
