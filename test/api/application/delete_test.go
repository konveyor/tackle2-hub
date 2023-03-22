package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationDelete(t *testing.T) {
	samples := CloneSamples()
	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			Create(t, r)

			// Try delete.
			err := Client.Delete(fmt.Sprintf("%s/%d", api.ApplicationsRoot, r.ID))
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check the it was deleted.
			err = Client.Get(fmt.Sprintf("%s/%d", api.ApplicationsRoot, r.ID), &r)
			if err == nil {
				t.Errorf("Exits, but should be deleted: %v", r)
			}
		})
	}
}
