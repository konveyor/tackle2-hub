package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

func TestApplicationList(t *testing.T) {
	samples := Samples()
	// Create.
	for i := range samples {
		c.Must(t, Create(&samples[i])) // Need modify parent array instead of single record created by the loop (to keep created IDs).
	}

	// Try list.
	got := []api.Application{}
	err := Client.Get(api.ApplicationsRoot, &got)
	if err != nil {
		t.Errorf("List error: %v", err.Error())
	}

	// Assert the response.
	if len(got) != len(samples) {
		t.Errorf("Wrong list length. Got %d, expected %d.", len(got), len(samples))
	}
	if c.FlatEqual(got, samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	// Clean.
	for _, r := range samples {
		c.Must(t, Delete(&r))
	}
}
