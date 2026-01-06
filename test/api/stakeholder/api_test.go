package stakeholder

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestStakeholderCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Stakeholder.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Stakeholder.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = Stakeholder.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Stakeholder.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = Stakeholder.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Stakeholder.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestStakeholderList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Stakeholder.Create(&sample))
		samples[name] = sample
	}

	got, err := Stakeholder.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Stakeholder.Delete(r.ID))
	}
}
