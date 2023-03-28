package jobfunction

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

func TestJobfunctionsCRUD(t *testing.T) {
	samples := Samples()

	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Client.Post(api.JobFunctionsRoot, &r)
			if err != nil {
				t.Errorf(err.Error())
			}
			rPath := client.Path(api.JobFunctionRoot, client.Params{api.ID: r.ID})

			// Get.
			got := api.JobFunction{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			if client.FlatEqual(got, &r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			updated := api.JobFunction{
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

func TestJobFunctionsList(t *testing.T) {
	samples := Samples()

	for _, r := range samples {
		Create(t, &r)
	}

	got := []api.Tag{}
	err := Client.Get(api.JobFunctionsRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if client.FlatEqual(got, samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		Delete(t, &r)
	}
}

func TestJobFunctionSeed(t *testing.T) {
	got := []api.JobFunction{}
	err := Client.Get(api.JobFunctionsRoot, &got)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(got) < 1 {
		t.Errorf("Seed looks empty, but it shouldn't.")
	}
}
