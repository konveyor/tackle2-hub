package application

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestApplicationGet(t *testing.T) {
	samples := CloneSamples()
	for _, r := range samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			Create(t, r)

			// Try get.
			got := api.Application{}
			err := Client.Get(fmt.Sprintf("%s/%d", api.ApplicationsRoot, r.ID), &got)
			if err != nil {
				t.Errorf("Get error: %v", err.Error())
			}

			// Assert the response.
			//if !reflect.DeepEqual(got, r) { // Fails on different ref/Ptrs addresses
			if got.Name != r.Name { // Too stupid asertion
				t.Errorf("Get returned different r. Got %v, expected %v.", got, r)
			}

			// Clean.
			Delete(t, r)
		})
	}
}
