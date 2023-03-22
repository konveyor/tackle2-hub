package stakeholdergroup

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

func TestStakeholderGroupCRUD(t *testing.T) {
	samples := CloneSamples()

	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.StakeholderGroupsRoot, &r)
			if err != nil {
				t.Errorf(err.Error())
			}
			rPath := fmt.Sprintf("%s/%d", api.StakeholderGroupsRoot, r.ID)

			// Get.
			got := api.StakeholderGroup{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if client.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			updated := api.StakeholderGroup{
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

func TestStakeholderGroupList(t *testing.T) {
	samples := CloneSamples()
	for _, r := range samples {
		Create(t, r)
	}

	got := []api.StakeholderGroup{}
	err := Client.Get(api.TagCategoriesRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if client.FlatEqual(got, samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		Delete(t, r)
	}

}
