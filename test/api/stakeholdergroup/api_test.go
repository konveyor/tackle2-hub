package stakeholdergroup

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestStakeholderGroupCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := StakeholderGroup.Create(&r)
			if err != nil {
				t.Error(err)
			}

			// Get.
			got, err := StakeholderGroup.Get(r.ID)
			if err != nil {
				t.Error(err)
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = StakeholderGroup.Update(&r)
			if err != nil {
				t.Error(err)
			}

			got, err = StakeholderGroup.Get(r.ID)
			if err != nil {
				t.Error(err)
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = StakeholderGroup.Delete(r.ID)
			if err != nil {
				t.Error(err)
			}

			_, err = StakeholderGroup.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestStakeholderGroupList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, StakeholderGroup.Create(&sample))
		samples[name] = sample
	}

	got, err := StakeholderGroup.List()
	if err != nil {
		t.Error(err)
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, StakeholderGroup.Delete(r.ID))
	}
}
