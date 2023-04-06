package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

func TestApplicationDelete(t *testing.T) {
	samples := Samples()
	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			c.Must(t, Create(&r))
			rPath := c.Path(api.ApplicationRoot, c.Params{api.ID: r.ID})

			// Try delete.
			err := Client.Delete(rPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check the it was deleted.
			err = Client.Get(rPath, &r)
			if err == nil {
				t.Errorf("Exits, but should be deleted: %v", r)
			}
		})
	}
}
