package application

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationUpdateName(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			assert.Must(t, Application.Create(&r))

			// Update.
			update := r
			update.Name = fmt.Sprint(r.Name, " updated")
			assert.Should(t, Application.Update(&update))

			// Check the updated.
			got := api.Application{}
			got.ID = r.ID
			assert.Should(t, Application.Get(&got))
			if !reflect.DeepEqual(got.Name, update.Name) {
				t.Errorf("Different updated name error. Got %v, expected %v", got.Name, update.Name)
			}

			// Clean.
			assert.Must(t, Application.Delete(&r))
		})
	}
}

// Tests updating different Applications attributes and references resources will be added here.
