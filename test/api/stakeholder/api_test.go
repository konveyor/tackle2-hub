package stakeholder

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

func TestStakeholderCRUD(t *testing.T) {
	samples := Samples()

	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.StakeholdersRoot, &r)
			if err != nil {
				t.Errorf(err.Error())
			}
			rPath := c.Path(api.StakeholderRoot, c.Params{api.ID: r.ID})

			// Get.
			got := api.Stakeholder{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if c.FlatEqual(got, &r) {
				t.Errorf("Different response error. Got %v, expected %v", got, &r)
			}

			// Update.
			update := api.Stakeholder{
				Name:  "Updated " + r.Name,
				Email: r.Email,
			}
			err = Client.Put(rPath, update)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != update.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, update.Name)
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

func TestStakeholderList(t *testing.T) {
	samples := Samples()
	for i := range samples {
		c.Must(t, Create(&samples[i]))
	}

	got := []api.Stakeholder{}
	err := Client.Get(api.StakeholdersRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if c.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, &samples)
	}

	for _, r := range samples {
		c.Must(t, Delete(&r))
	}

}
