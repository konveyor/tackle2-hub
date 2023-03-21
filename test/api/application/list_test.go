package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationList(t *testing.T) {
	samples := Samples()
	// Create applications.
	for _, application := range samples {
		Create(t, application)
	}

	// Try list applications.
	gotApplications := []*api.Application{}
	err := Client.Get(api.ApplicationsRoot, &gotApplications)
	if err != nil {
		t.Errorf("List error: %v", err.Error())
	}

	// Assert the response.
	if len(gotApplications) != len(samples) {
		t.Errorf("Wrong list length. Got %d, expected %d.", len(gotApplications), len(samples))
	}

	// Reflect fails on different Ptr address, compare via json serialization?
	//if !reflect.DeepEqual(gotApplications, Samples) {
	//	t.Errorf("List returned different applications. Got %v, expected %v.", &gotApplications, &Samples)
	//}

	// Clean the application.
	for _, application := range samples {
		Delete(t, application)
	}
}
