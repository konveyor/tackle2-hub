package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationGet(t *testing.T) {
	samples := Samples()
	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			assert.Must(t, Create(&r))

			// Try get.
			got := api.Application{}
			err := Client.Get(client.Path(api.ApplicationRoot, client.Params{api.ID: r.ID}), &got)
			if err != nil {
				t.Errorf("Get error: %v", err.Error())
			}

			// Assert the response.
			if assert.FlatEqual(got, samples) {
				t.Errorf("Different response error. Got %v, expected %v", got, samples)
			}

			// Clean.
			assert.Must(t, Delete(&r))
		})
	}
}
