package tagcategory

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

func TestTagCategoriesCRUD(t *testing.T) {
	samples := CloneSamples()

	for _, tagCategory := range samples {
		t.Run(tagCategory.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.TagCategoriesRoot, &tagCategory)
			if err != nil {
				t.Errorf(err.Error())
			}
			tagCategoryPath := fmt.Sprintf("%s/%d", api.TagCategoriesRoot, tagCategory.ID)

			// Get.
			gotTagCategory := api.TagCategory{}
			err = Client.Get(tagCategoryPath, &gotTagCategory)
			if err != nil {
				t.Errorf(err.Error())
			}
			if client.FlatEqual(gotTagCategory, tagCategory) {
				t.Errorf("Different response error. Got %v, expected %v", gotTagCategory, tagCategory)
			}

			// Update.
			updatedTagCategory := api.TagCategory{
				Name: "Updated " + tagCategory.Name,
			}
			err = Client.Put(tagCategoryPath, updatedTagCategory)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(tagCategoryPath, &gotTagCategory)
			if err != nil {
				t.Errorf(err.Error())
			}
			if gotTagCategory.Name != updatedTagCategory.Name {
				t.Errorf("Different response error. Got %s, expected %s", gotTagCategory.Name, updatedTagCategory.Name)
			}

			// Delete.
			err = Client.Delete(tagCategoryPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(tagCategoryPath, &tagCategory)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", tagCategory)
			}

			// Cleanup.
			Delete(t, tagCategory)
		})
	}
}

func TestTagCategoriesList(t *testing.T) {
	samples := CloneSamples()
	for _, tagCategory := range samples {
		Create(t, tagCategory)
	}

	gotTagCategories := []api.TagCategory{}
	err := Client.Get(api.TagCategoriesRoot, &gotTagCategories)
	if err != nil {
		t.Errorf(err.Error())
	}
	if client.FlatEqual(gotTagCategories, samples) {
		t.Errorf("Different response error. Got %v, expected %v", gotTagCategories, samples)
	}

	for _, tagCategory := range samples {
		Delete(t, tagCategory)
	}

}
