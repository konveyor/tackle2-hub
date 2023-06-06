package dependency

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestDependencyCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("Dependency_CRUD", func(t *testing.T) {

			applicationFrom := api.Application{
				Name:        r.From.Name,
				Description: "From application",
			}

			applicationTo := api.Application{
				Name:        r.To.Name,
				Description: "To application",
			}

			// Create.
			dependency := api.Dependency{
				From: api.Ref{
					ID: applicationFrom.ID,
				},
				To: api.Ref{
					ID: applicationTo.ID,
				},
			}
			err := Dependency.Create(&dependency)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Dependency.Get(dependency.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, dependency.ID) {
				t.Errorf("Different response error. Got %v, expected %v", got, dependency.ID)
			}

			_, err = Dependency.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Delete dependency.
			err = Dependency.Delete(dependency.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestDependencyList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]

		// Create applications.
		applicationFrom := api.Application{
			Name:        sample.From.Name,
			Description: "From application",
		}
		sample.From.ID = applicationFrom.ID

		applicationTo := api.Application{
			Name:        sample.To.Name,
			Description: "To application",
		}
		sample.To.ID = applicationTo.ID

		samples[name] = sample
	}

	// List dependencies.
	got, err := Dependency.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	// Delete dependencies
	for _, r := range samples {
		assert.Must(t, Dependency.Delete(r.ID))
	}
}
