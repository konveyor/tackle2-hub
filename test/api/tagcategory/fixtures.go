package tagcategory

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
)

var (
	// Setup Hub API client
	Client = client.Client
)

//
// Set of valid TagCategories resources for tests and reuse.
var SampleTagCategories = []*api.TagCategory{
	{
		Name:  "Test OS",
		Color: "#dd0000",
	},
	{
		Name:  "Test Language",
		Color: "#0000dd",
	},
}

//
// Creates a copy of Samples for given test (copy is there to avoid tests inflence each other using the same object ref).
func CloneSamples() (samples []*api.TagCategory) {
	raw, err := json.Marshal(SampleTagCategories)
	if err != nil {
		panic("ERROR cloning samples")
	}
	json.Unmarshal(raw, &samples)
	return
}

//
// Create.
func Create(t *testing.T, r *api.TagCategory) {
	err := Client.Post(api.TagCategoriesRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error())
	}
}

//
// Delete.
func Delete(t *testing.T, r *api.TagCategory) {
	err := Client.Delete(fmt.Sprintf("%s/%d", api.TagCategoriesRoot, r.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
