package tagcategory

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

func TestTagCategoriesCRUD(t *testing.T) {
	samples := Samples()

	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.TagCategoriesRoot, &r)
			if err != nil {
				t.Errorf(err.Error())
			}
			rPath := c.Path(api.TagCategoryRoot, c.Params{api.ID: r.ID})

			// Get.
			got := api.TagCategory{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if c.FlatEqual(got, &r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			updated := api.TagCategory{
				Name: "Updated " + r.Name,
			}
			err = Client.Put(rPath, updated)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != updated.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, updated.Name)
			}

			// Delete.
			err = Client.Delete(rPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(rPath, &r)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestTagCategoriesList(t *testing.T) {
	samples := Samples()
	for i := range samples {
		c.Must(t, Create(&samples[i]))
	}

	got := []api.TagCategory{}
	err := Client.Get(api.TagCategoriesRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if c.FlatEqual(got, samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		c.Must(t, Delete(&r))
	}

}

func TestTagCategorySeed(t *testing.T) {
	got := []api.TagCategory{}
	err := Client.Get(api.TagCategoriesRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(got) < 1 {
		t.Errorf("Seed looks empty, but it shouldn't.")
	}
}
