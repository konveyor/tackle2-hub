package stakeholdergroup

import (
	"testing"

	assert2 "github.com/konveyor/tackle2-hub/internal/test/assert"
)

func TestStakeholderGroupCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := StakeholderGroup.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := StakeholderGroup.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert2.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = StakeholderGroup.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = StakeholderGroup.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = StakeholderGroup.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
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
		assert2.Must(t, StakeholderGroup.Create(&sample))
		samples[name] = sample
	}

	got, err := StakeholderGroup.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert2.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert2.Must(t, StakeholderGroup.Delete(r.ID))
	}
}
