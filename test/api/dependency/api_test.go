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

			applicationFrom := sample.ApplicationFrom
			assert.Must(t, Application.Create(&applicationFrom))

			applicationTo := sample.ApplicationTo
			assert.Must(t, Application.Create(&applicationTo))

			// Create.
			dependency := api.Dependency{
				From: api.Ref{
					ID: applicationFrom.ID,
				},
				To: api.Ref{
					ID: applicationTo.ID,
				},
			}
			assert.Must(t, Dependency.Create(&dependency))

			// Get.
			got, err := Dependency.Get(dependency.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, dependency.ID) {
				t.Errorf("Different response error. Got %v, expected %v", got, dependency.ID)
			}

			// Delete dependency.
			assert.Must(t, Dependency.Delete(dependency.ID))

			//Delete Applications
			assert.Must(t, Application.Delete(applicationFrom.ID))
			assert.Must(t, Application.Delete(applicationTo.ID))
		})
	}
}

func TestDependencyList(t *testing.T) {
	for _, sample := range Samples {
		// Create applications.
		applicationFrom := sample.ApplicationFrom
		assert.Must(t, Application.Create(&applicationFrom))
		sample.ApplicationFrom.ID = applicationFrom.ID

		applicationTo := sample.ApplicationTo
		assert.Must(t, Application.Create(&applicationTo))
		sample.ApplicationTo.ID = applicationTo.ID
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
		assert.Must(t, Application.Delete(sample.ApplicationFrom.ID))
		assert.Must(t, Application.Delete(sample.ApplicationTo.ID))
	}
}
