package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	c "github.com/konveyor/tackle2-hub/test/api/client"
)

func TestApplicationGet(t *testing.T) {
	samples := Samples()
	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			c.Must(t, Create(&r))

			// Try get.
			got := api.Application{}
			err := Client.Get(c.Path(api.ApplicationRoot, c.Params{api.ID: r.ID}), &got)
			if err != nil {
				t.Errorf("Get error: %v", err.Error())
			}

			// Assert the response.
			if c.FlatEqual(got, samples) {
				t.Errorf("Different response error. Got %v, expected %v", got, samples)
			}

			// Clean.
			c.Must(t, Delete(&r))
		})
	}
}
