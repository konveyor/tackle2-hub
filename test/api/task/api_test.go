package task

import (
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/filter"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTaskCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Task.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Task.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = Task.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Task.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// patch.
			type TaskPatch struct {
				Name string `json:"name"`
			}
			p := &TaskPatch{}
			p.Name = "patched " + r.Name
			err = Task.Patch(r.ID, p)
			if err != nil {
				t.Errorf(err.Error())
			}
			got, err = Task.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != p.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, p.Name)
			}
			// Delete.
			err = Task.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			for i := 5; i >= 0; i-- {
				time.Sleep(time.Second)
				_, err = Task.Get(r.ID)
				if err != nil {
					break
				}
				if i == 0 {
					t.Errorf("Resource exits, but should be deleted: %v", r)
					break
				}
			}
		})
	}
}

func TestTaskList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Task.Create(&sample))
		samples[name] = sample
	}

	got, err := Task.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Task.Delete(r.ID))
	}
}

func TestBulkCancel(t *testing.T) {
	f := binding.Filter{}
	f.And("id").Eq(filter.Any{1, 2, 3})
	err := Task.BulkCancel(f)
	if err != nil {
		t.Errorf(err.Error())
	}
}
