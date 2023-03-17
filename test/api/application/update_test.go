package application

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationUpdateName(t *testing.T) {
	samples := Samples()
	for _, application := range samples {
		t.Run(fmt.Sprintf("Update application %s", application.Name), func(t *testing.T) {
			// Create the application.
			Create(t, application)

			// Update the application.
			updatedName := fmt.Sprint(application.Name, " updated")
			updateApplication := api.Application{
				Name: updatedName,
			}
			err = Client.Put(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID), &updateApplication)
			if err != nil {
				t.Errorf("Update error: %v", err.Error())
			}

			// Check the updated application.
			gotApplication := api.Application{}
			err = Client.Get(fmt.Sprintf("%s/%d", api.ApplicationsRoot, application.ID), &gotApplication)
			if err != nil {
				t.Errorf("Get updated error: %v", err.Error())
			}
			if !reflect.DeepEqual(gotApplication.Name, updateApplication.Name) {
				t.Errorf("Different updated name error. Got %v, expected %v", gotApplication.Name, updateApplication.Name)
			}

			// Clean the application.
			Delete(t, application)
		})
	}
}
