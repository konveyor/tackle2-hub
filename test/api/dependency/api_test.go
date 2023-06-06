package dependency

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestDependencyCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("Dependency CRUD", func(t *testing.T) {
			// Create.
			err := Dependency.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Dependency.Get(r.From.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.From.Name = "Updated " + r.From.Name
			err = Dependency.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Dependency.Get(r.From.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.From.Name != r.From.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.From.Name, r.From.Name)
			}

			// Delete.
			err = Dependency.Delete(r.From.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Dependency.Get(r.From.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestDependencyList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Dependency.Create(&sample))
		samples[name] = sample
	}

	got, err := Dependency.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Dependency.Delete(r.ID))
	}
}
