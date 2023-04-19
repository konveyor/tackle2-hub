package task

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTaskCRUD(t *testing.T) {
	samples := Samples()

	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.TasksRoot, &r)
			if err != nil {
				t.Errorf(err.Error())
			}
			rPath := client.Path(api.TaskRoot, client.Params{api.ID: r.ID})

			// Get.
			got := api.Task{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, &r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			updated := api.Task{
				Name:  "Updated " + r.Name,
				Addon: "updated-" + r.Addon,
				Data:  "{updated}",
				State: "Created",
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

func TestTaskList(t *testing.T) {
	samples := Samples()

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Create(&sample))
		samples[name] = sample
	}

	got := []api.Task{}
	err := Client.Get(api.TasksRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Delete(&r))
	}
}
