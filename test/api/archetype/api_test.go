package archetype

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestArchetypeCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Archetype.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Archetype.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = Archetype.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Archetype.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = Archetype.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Archetype.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestArchetypeList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Archetype.Create(&sample))
		samples[name] = sample
	}

	got, err := Archetype.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Archetype.Delete(r.ID))
	}
}

// TODO(maufart): Add assessment POST&GET methods
