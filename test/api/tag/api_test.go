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

	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.TagsRoot, &r)
			if err != nil {
				t.Errorf(err.Error())
			}
			rPath := fmt.Sprintf("%s/%d", api.TagsRoot, r.ID)

			// Get.
			got := api.Tag{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if client.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			updated := api.Tag{
				Name:     "Updated " + r.Name,
				Category: r.Category,
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

			// Cleanup.
			Delete(t, r)
		})
	}
	// Category cleanup.
	tagcategory.Delete(t, category)

}

func TestTagList(t *testing.T) {
	samples := CloneSamples()
	category := prepareCategory(t, samples)

	for _, r := range samples {
		Create(t, r)
	}

	got := []api.Tag{}
	err := Client.Get(api.TagsRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if client.FlatEqual(got, samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		Delete(t, r)
	}
	tagcategory.Delete(t, category)
}
