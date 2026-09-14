package tagcategory

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTagCategoryCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			var err error
			// Create.
			err = TagCategory.Create(&r)
			if err != nil {
				t.Error(err)
			}

			// Get.
			got, err := TagCategory.Get(r.ID)
			if err != nil {
				t.Error(err)
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = TagCategory.Update(&r)
			if err != nil {
				t.Error(err)
			}

			got, err = TagCategory.Get(r.ID)
			if err != nil {
				t.Error(err)
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = TagCategory.Delete(r.ID)
			if err != nil {
				t.Error(err)
			}
			_, err = TagCategory.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}

			// Ensure.
			r.ID = 0
			err = TagCategory.Ensure(&r)
			if err != nil {
				t.Error(err)
			}
			got, err = TagCategory.Get(r.ID)
			if err != nil {
				t.Error(err)
			}
			if got.ID == 0 {
				t.Errorf("Ensured resource has no id.")
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}
			err = TagCategory.Delete(r.ID)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestTagCategoryList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, TagCategory.Create(&sample))
		samples[name] = sample
	}

	got, err := TagCategory.List()
	if err != nil {
		t.Error(err)
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, TagCategory.Delete(r.ID))
	}
}

func TestTagCategorySeed(t *testing.T) {
	got, err := TagCategory.List()
	if err != nil {
		t.Error(err)
	}
	if len(got) < 1 {
		t.Errorf("Seed looks empty, but it shouldn't.")
	}
}
