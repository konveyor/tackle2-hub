package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationList(t *testing.T) {
	samples := CloneSamples()
	// Create.
	for _, r := range samples {
		Create(t, r)
	}

	// Try list.
	got := []*api.Application{}
	err := Client.Get(api.ApplicationsRoot, &got)
	if err != nil {
		t.Errorf("List error: %v", err.Error())
	}

	// Assert the response.
	if len(got) != len(samples) {
		t.Errorf("Wrong list length. Got %d, expected %d.", len(got), len(samples))
	}

	// Reflect fails on different Ptr address, compare via json serialization?
	//if !reflect.DeepEqual(got, Samples) {
	//	t.Errorf("List returned different applications. Got %v, expected %v.", &got, &Samples)
	//}

	// Clean.
	for _, r := range samples {
		Delete(t, r)
	}
}
