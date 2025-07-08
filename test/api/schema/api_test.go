package schema

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestGetSchema(t *testing.T) {
	// Get.
	api := api.RestAPI{}
	err := RichClient.Client.Get("/schema", &api)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(api.Routes) < 1 {
		t.Errorf("Got empty Paths from /schema.")
	}
}
