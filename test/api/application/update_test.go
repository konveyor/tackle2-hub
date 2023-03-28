package application

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

func TestApplicationUpdateName(t *testing.T) {
	samples := Samples()
	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			Create(t, &r)
			rPath := client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID})

			// Update.
			updatedName := fmt.Sprint(r.Name, " updated")
			update := api.Application{
				Name: updatedName,
			}
			err := Client.Put(rPath, &update)
			if err != nil {
				t.Errorf("Update error: %v", err.Error())
			}

			// Check the updated.
			got := api.Application{}
			err = Client.Get(rPath, &got)
			if err != nil {
				t.Errorf("Get updated error: %v", err.Error())
			}
			if !reflect.DeepEqual(got.Name, update.Name) {
				t.Errorf("Different updated name error. Got %v, expected %v", got.Name, update.Name)
			}

			// Clean.
			Delete(t, &r)
		})
	}
}
