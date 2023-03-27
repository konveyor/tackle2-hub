package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

var SampleFacts = []*api.Fact{
	{
		Key:   "pet",
		Value: "{\"kind\":\"dog\",\"Age\":4}",
	},
	{
		Key:   "address",
		Value: "{\"street\":\"Maple\",\"State\":\"AL\"}",
	},
}

func TestApplicationFactCRUD(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	Create(t, &application)

	// Test Facts subresource.
	for _, r := range SampleFacts {
		t.Run(fmt.Sprintf("Fact %s application %s", r.Key, application.Name), func(t *testing.T) {

			factPath := fmt.Sprintf("%s/%d/facts/%s", api.ApplicationsRoot, application.ID, r.Key)

			// Create.
			err := Client.Post(factPath, &r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got := api.Fact{}
			err = Client.Get(factPath, &got)
			if err != nil {
				t.Errorf(err.Error())
			}
			// Not sure about map[] wrapping of the Fact Value (interface) by the API
			//if !reflect.DeepEqual(got.Value, fact.Value) {
			//	t.Errorf("Different fact value error. Got %v, expected %v", got.Value, fact.Value)
			//}

			// Update.
			updated := api.Fact{
				Value: fmt.Sprintf("{\"%s\":\"%s\"}", r.Key, "updated"),
			}
			err = Client.Put(factPath, updated)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get the updated.
			got = api.Fact{}
			err = Client.Get(factPath, &got)
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
			err = Client.Get(factPath, &got)
			if err == nil {
				t.Errorf("Exits, but should be deleted: %v", r)
			}
		})
	}

	// Clean the application.
	Delete(t, &application)
}

func TestApplicationFactsList(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	Create(t, &application)

	// Create facts.
	for _, r := range SampleFacts {
		err := Client.Post(fmt.Sprintf("%s/%d/facts/%s", api.ApplicationsRoot, application.ID, r.Key), &r)
		if err != nil {
			t.Fatalf(err.Error())
		}
	}

	// Check facts list with and without trailing slash (client maybe removes it anyway).
	factsPathSuffix := []string{"facts", "facts/"}
	for _, pathSuffix := range factsPathSuffix {
		t.Run(fmt.Sprintf("Fact list application %s with %s", application.Name, pathSuffix), func(t *testing.T) {
			got := []api.Fact{}
			err := Client.Get(fmt.Sprintf("%s/%d/%s", api.ApplicationsRoot, application.ID, pathSuffix), &got)
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
	Delete(t, &application)
}
