package dependency

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	app "github.com/konveyor/tackle2-hub/test/api/application"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestDependencyCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(fmt.Sprintf("Dependency from %s -> %s", r.From.Name, r.To.Name), func(t *testing.T) {

			applicationFrom := api.Application{
				Name:        r.From.Name,
				Description: "From application",
			}

			err := app.Application.Create(&applicationFrom)
			if err != nil {
				t.Errorf(err.Error())
			}

			applicationTo := api.Application{
				Name:        r.To.Name,
				Description: "To application",
			}

			err = app.Application.Create(&applicationTo)
			if err != nil {
				t.Errorf(err.Error())
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

			err = Dependency.Create(&dependency)
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

			// Delete dependency.
			err = Dependency.Delete(dependency.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			//Delete Applications
			err = app.Application.Delete(applicationFrom.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = app.Application.Delete(applicationTo.ID)
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
		assert.Must(t, app.Application.Create(&applicationFrom))
		sample.From.ID = applicationFrom.ID

		applicationTo := api.Application{
			Name:        sample.To.Name,
			Description: "To application",
		}
		assert.Must(t, app.Application.Create(&applicationTo))
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

	// Delete dependencies as well as applications.
	for _, r := range samples {
		assert.Must(t, Dependency.Delete(r.ID))
		assert.Must(t, app.Application.Delete(r.From.ID))
		assert.Must(t, app.Application.Delete(r.To.ID))
	}
}
