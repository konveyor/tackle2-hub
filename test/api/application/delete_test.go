package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationDelete(t *testing.T) {
	samples := Samples()
	for _, application := range samples {
		t.Run(fmt.Sprintf("Delete application %s", application.Name), func(t *testing.T) {
			// Create the application.
			Create(t, application)

			// Delete the application.
			err := Client.Delete(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID))
			if err != nil {
				t.Errorf("Delete error: %v", err.Error())
			}

			// Check the application was deleted.
			err = Client.Get(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID), &application)
			if err == nil {
				t.Errorf("Application exits, but should be deleted: %v", application)
			}
		})
	}
}
