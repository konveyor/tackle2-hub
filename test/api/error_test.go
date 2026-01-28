package api

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

func TestErrors(t *testing.T) {
	richClient := client.PrepareRichClient()
	_, err := richClient.Application.Get(0)
	if !errors.Is(err, &api.NotFound{}) {
		t.Fatalf("Expecting NotFound")
	}
}
