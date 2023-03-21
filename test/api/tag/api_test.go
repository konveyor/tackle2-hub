package tag

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/api/tagcategory"
)

func TestTagCRUD(t *testing.T) {
	samples := CloneSamples()
	category := prepareCategory(t, samples)

	for _, tag := range samples {
		t.Run(tag.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.TagsRoot, &tag)
			if err != nil {
				t.Errorf(err.Error())
			}
			tagPath := fmt.Sprintf("%s/%d", api.TagsRoot, tag.ID)

			// Get.
			gotTag := api.Tag{}
			err = Client.Get(tagPath, &gotTag)
			if err != nil {
				t.Errorf(err.Error())
			}
			if client.FlatEqual(gotTag, tag) {
				t.Errorf("Different response error. Got %v, expected %v", gotTag, tag)
			}

			// Update.
			updatedTag := api.Tag{
				Name:     "Updated " + tag.Name,
				Category: tag.Category,
			}
			err = Client.Put(tagPath, updatedTag)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(tagPath, &gotTag)
			if err != nil {
				t.Errorf(err.Error())
			}
			if gotTag.Name != updatedTag.Name {
				t.Errorf("Different response error. Got %s, expected %s", gotTag.Name, updatedTag.Name)
			}

			// Delete.
			err = Client.Delete(tagPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(tagPath, &tag)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", tag)
			}

			// Cleanup.
			Delete(t, tag)
		})
	}
	// Category cleanup.
	tagcategory.Delete(t, category)

}

func TestTagList(t *testing.T) {
	samples := CloneSamples()
	category := prepareCategory(t, samples)

	for _, tag := range samples {
		Create(t, tag)
	}

	gotTag := []api.Tag{}
	err := Client.Get(api.TagsRoot, &gotTag)
	if err != nil {
		t.Errorf(err.Error())
	}
	if client.FlatEqual(gotTag, samples) {
		t.Errorf("Different response error. Got %v, expected %v", gotTag, samples)
	}

	for _, tag := range samples {
		Delete(t, tag)
	}
	tagcategory.Delete(t, category)
}
