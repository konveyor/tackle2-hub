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

			assert.Should(t, Application.Create(&sample.ApplicationFrom))
			assert.Should(t, Application.Create(&sample.ApplicationTo))

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
			if assert.FlatEqual(got, dependency.ID) {
				t.Errorf("Different response error. Got %v, expected %v", got, dependency.ID)
			}

			// Delete dependency.
			assert.Should(t, Dependency.Delete(dependency.ID))

			//Delete Applications
			assert.Should(t, Application.Delete(sample.ApplicationFrom.ID))
			assert.Should(t, Application.Delete(sample.ApplicationTo.ID))
		})
	}
}

func TestDependencyList(t *testing.T) {
	for _, sample := range Samples {

		// Create applications.
		assert.Should(t, Application.Create(&sample.ApplicationFrom))
		assert.Should(t, Application.Create(&sample.ApplicationTo))
	}

	// List dependencies.
	got, err := Dependency.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, Samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, Samples)
	}

	// Delete applications.
	for _, sample := range Samples {
		assert.Should(t, Application.Delete(sample.ApplicationFrom.ID))
		assert.Should(t, Application.Delete(sample.ApplicationTo.ID))
	}
}
