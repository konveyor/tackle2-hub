package application

import (
	"fmt"
	"testing"

	api2 "github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

var SampleFacts = []*api2.Fact{
	{
		Key:    "pet",
		Value:  "{\"kind\":\"dog\",\"Age\":4}",
		Source: "test",
	},
	{
		Key:    "address",
		Value:  "{\"street\":\"Maple\",\"State\":\"AL\"}",
		Source: "test",
	},
}

func TestApplicationFactCRUD(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Minimal

	// Create the application.
	assert.Must(t, Application.Create(&application))

	// Test Facts subresource.
	for _, r := range SampleFacts {
		t.Run(fmt.Sprintf("Fact %s application %s", r.Key, application.Name), func(t *testing.T) {
			key := api2.FactKey(r.Key)
			key.Qualify(r.Source)
			factPath := binding.Path(api2.ApplicationFactRoute).Inject(binding.Params{api2.ID: application.ID, api2.Key: key})

			// Create.
			err := Client.Post(binding.Path(api2.ApplicationFactsRoute).Inject(binding.Params{api2.ID: application.ID}), &r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			var v any
			err = Client.Get(factPath, &v)
			if err != nil {
				t.Errorf(err.Error())
			}
			// Not sure about map[] wrapping of the Fact Value (interface) by the API
			//if !reflect.DeepEqual(got.Value, fact.Value) {
			//	t.Errorf("Different fact value error. Got %v, expected %v", got.Value, fact.Value)
			//}

			// Update.
			updated := api2.Fact{
				Value: fmt.Sprintf("{\"%s\":\"%s\"}", r.Key, "updated"),
			}
			err = Client.Put(factPath, updated.Value)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get the updated.
			err = Client.Get(factPath, &v)
			if err != nil {
				t.Errorf(err.Error())
			}
			//if !reflect.DeepEqual(got.Value, updated.Value) {
			//	t.Errorf("Different updated fact value error. Got %v, expected %v", got.Value, updated.Value)
			//}

			// Delete.
			err = Client.Delete(factPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check the it was deleted.
			err = Client.Get(factPath, &v)
			if err == nil {
				t.Errorf("Exits, but should be deleted: %v", r)
			}
		})
	}

	// Clean the application.
	assert.Must(t, Application.Delete(application.ID))
}

func TestApplicationFactsList(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Minimal

	// Create the application.
	assert.Must(t, Application.Create(&application))

	// Create facts.
	for _, r := range SampleFacts {
		err := Client.Post(binding.Path(api2.ApplicationFactsRoute).Inject(binding.Params{api2.ID: application.ID}), &r)
		if err != nil {
			t.Errorf(err.Error())
		}
	}

	// Check facts list with and without trailing slash (client maybe removes it anyway).
	factsPathSuffix := []string{"facts/test:", "facts/test:/"}
	for _, pathSuffix := range factsPathSuffix {
		t.Run(fmt.Sprintf("Fact list application %s with %s", application.Name, pathSuffix), func(t *testing.T) {
			got := api2.Map{}
			err := Client.Get(fmt.Sprintf("%s/%s", binding.Path(api2.ApplicationRoute).Inject(binding.Params{api2.ID: application.ID}), pathSuffix), &got)
			if err != nil {
				t.Errorf("Get list error: %v", err.Error())
			}
			if len(got) != len(SampleFacts) {
				t.Errorf("Different length of fact list error. Got %d, expected %d", len(got), len(SampleFacts))
			}
			// Compare returned list values?
		})
	}

	// Clean the application.
	assert.Must(t, Application.Delete(application.ID))
}
