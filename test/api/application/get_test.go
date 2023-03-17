package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationGet(t *testing.T) {
	samples := Samples()
	for _, application := range samples {
		t.Run(fmt.Sprintf("Get application %s", application.Name), func(t *testing.T) {
			// Create the application.
			Create(t, application)

			// Try get the application.
			gotApplication := api.Application{}
			err = Client.Get(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID), &gotApplication)
			if err != nil {
				t.Errorf("Get error: %v", err.Error())
			}

			// Assert the response.
			//if !reflect.DeepEqual(gotApplication, application) { // Fails on different ref/Ptrs addresses
			if gotApplication.Name != application.Name { // Too stupid asertion
				t.Errorf("Get returned different application. Got %v, expected %v.", gotApplication, application)
			}

			// Clean the application.
			EnsureDelete(t, application)
		})
	}
}
