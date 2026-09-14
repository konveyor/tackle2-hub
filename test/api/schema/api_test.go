package schema

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
)

const (
	name = "cloudfoundry-coordinates"
)

func TestGetSchema(t *testing.T) {
	// Get.
	api := api.RestAPI{}
	err := RichClient.Client.Get("/schema", &api)
	if err != nil {
		t.Error(err)
	}
	if len(api.Routes) < 1 {
		t.Errorf("Got empty Paths from /schema.")
	}
}

func TestGet(t *testing.T) {
	r, err := RichClient.Schema.Get(name)
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != name {
		t.Errorf("Name: '%s' expected.", name)
	}
}

func TestFind(t *testing.T) {
	r, err := RichClient.Schema.Find("platform", "cloudfoundry", "coordinates")
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != name {
		t.Errorf("Name: '%s' expected.", name)
	}
}
