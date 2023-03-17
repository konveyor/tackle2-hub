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
	Create(t, application)

	for _, fact := range SampleFacts {
		t.Run(fmt.Sprintf("Fact %s application %s", fact.Key, application.Name), func(t *testing.T) {

			factPath := fmt.Sprintf("%s/%d/facts/%s", api.ApplicationsRoot, application.ID, fact.Key)

			//fmt.Printf("#### fresh fact: %v", fact)
			// Create fact.
			err = Client.Post(factPath, &fact)
			if err != nil {
				t.Errorf("Create error: %v", err.Error())
			}
			//fmt.Printf("#### created fact: %v", fact)

			// Get the fact.
			gotFact := api.Fact{}
			err = Client.Get(factPath, &gotFact)
			if err != nil {
				t.Errorf("Get error: %v", err.Error())
			}
			// Not sure about map[] wrapping of the Fact Value (interface) by the API
			//if !reflect.DeepEqual(gotFact.Value, fact.Value) {
			//	t.Errorf("Different fact value error. Got %v, expected %v", gotFact.Value, fact.Value)
			//}

			// Update the fact.
			updatedFact := api.Fact{
				Value: fmt.Sprintf("{\"%s\":\"%s\"}", fact.Key, "updated"),
			}
			err = Client.Put(factPath, updatedFact)
			if err != nil {
				t.Errorf("Update error: %v", err.Error())
			}

			// Get the updated fact.
			gotFact = api.Fact{}
			err = Client.Get(factPath, &gotFact)
			if err != nil {
				t.Errorf("Get updated error: %v", err.Error())
			}
			//if !reflect.DeepEqual(gotFact.Value, updatedFact.Value) {
			//	t.Errorf("Different updated fact value error. Got %v, expected %v", gotFact.Value, updatedFact.Value)
			//}

			// Delete the fact.
			err = Client.Delete(factPath)
			if err != nil {
				t.Errorf("Delete error: %v", err.Error())
			}

			// Check the fact was deleted.
			err = Client.Get(factPath, &gotFact)
			if err == nil {
				t.Errorf("Application exits, but should be deleted: %v", application)
			}
		})
	}

	// Clean the application.
	EnsureDelete(t, application)
}

func TestApplicationFactsList(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	Create(t, application)

	// Create facts.
	for _, fact := range SampleFacts {
		err = Client.Post(fmt.Sprintf("%s/%d/facts/%s", api.ApplicationsRoot, application.ID, fact.Key), &fact)
		if err != nil {
			t.Fatalf("Create error: %v", err.Error())
		}
	}

	// Check facts list with and without trailing slash (client maybe removes it anyway).
	factsPathSuffix := []string{"facts", "facts/"}
	for _, pathSuffix := range factsPathSuffix {
		t.Run(fmt.Sprintf("Fact list application %s with %s", application.Name, pathSuffix), func(t *testing.T) {
			gotFacts := []api.Fact{}
			err = Client.Get(fmt.Sprintf("%s/%d/%s", api.ApplicationsRoot, application.ID, pathSuffix), &gotFacts)
			if err != nil {
				t.Errorf("Get list error: %v", err.Error())
			}
			if len(gotFacts) != len(SampleFacts) {
				t.Errorf("Different length of fact list error. Got %d, expected %d", len(gotFacts), len(SampleFacts))
			}
			// Compare returned list values?
		})
	}

	// Clean the application.
	EnsureDelete(t, application)
}
