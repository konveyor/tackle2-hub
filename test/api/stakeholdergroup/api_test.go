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
			gotR := api.StakeholderGroup{}
			err = Client.Get(rPath, &gotR)
			if err != nil {
				t.Errorf(err.Error())
			}
			if client.FlatEqual(gotR, r) {
				t.Errorf("Different response error. Got %v, expected %v", gotR, r)
			}

			// Update.
			updatedR := api.StakeholderGroup{
				Name: "Updated " + r.Name,
			}
			err = Client.Put(rPath, updatedR)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(rPath, &gotR)
			if err != nil {
				t.Errorf(err.Error())
			}
			if gotR.Name != updatedR.Name {
				t.Errorf("Different response error. Got %s, expected %s", gotR.Name, updatedR.Name)
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
}

func TestStakeholderGroupList(t *testing.T) {
	samples := CloneSamples()
	for _, r := range samples {
		Create(t, r)
	}

	gotList := []api.StakeholderGroup{}
	err := Client.Get(api.TagCategoriesRoot, &gotList)
	if err != nil {
		t.Errorf(err.Error())
	}
	if client.FlatEqual(gotList, samples) {
		t.Errorf("Different response error. Got %v, expected %v", gotList, samples)
	}

	for _, r := range samples {
		Delete(t, r)
	}

}
