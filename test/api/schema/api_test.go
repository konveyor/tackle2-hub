package schema

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestGetSchema(t *testing.T) {
	// Get.
	schema := api.Schema{}
	err := RichClient.Client.Get("/schema", &schema)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(schema.Paths) < 1 {
		t.Errorf("Got empty Paths from /schema.")
	}
}
