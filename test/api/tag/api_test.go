package tag

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTagCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			var err error
			// Create.
			err = Tag.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Tag.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = Tag.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Tag.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = Tag.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			_, err = Tag.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
			// Ensure.
			r.ID = 0
			err = Tag.Ensure(&r)
			if err != nil {
				t.Errorf(err.Error())
			}
			got, err = Tag.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.ID == 0 {
				t.Errorf("Ensured resource has no id.")
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}
			err = Tag.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestTagList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Tag.Create(&sample))
		samples[name] = sample
	}

	got, err := Tag.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Tag.Delete(r.ID))
	}
}


func TestTagSeed(t *testing.T) {
	got, err := Tag.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(got) < 1 {
		t.Errorf("Seed looks empty, but it shouldn't.")
	}
}