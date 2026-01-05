package jobfunction

import (
	"testing"

	assert2 "github.com/konveyor/tackle2-hub/internal/test/assert"
)

func TestJobFunctionCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := JobFunction.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := JobFunction.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert2.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = JobFunction.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = JobFunction.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = JobFunction.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = JobFunction.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestJobFunctionList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert2.Must(t, JobFunction.Create(&sample))
		samples[name] = sample
	}

	got, err := JobFunction.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert2.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert2.Must(t, JobFunction.Delete(r.ID))
	}
}

func TestJobFunctionSeed(t *testing.T) {
	got, err := JobFunction.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(got) < 1 {
		t.Errorf("Seed looks empty, but it shouldn't.")
	}
}
