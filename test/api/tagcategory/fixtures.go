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
		//Tags:  []api.Ref{},
	},
	{
		Name:  "Test Language",
		Color: "#0000dd",
		//Tags:  []api.Ref{},
	},
}

//
// Creates a copy of Samples for given test (copy is there to avoid tests inflence each other using the same object ref).
func CloneSamples() (applications []*api.TagCategory) {
	raw, err := json.Marshal(SampleTagCategories)
	if err != nil {
		fmt.Print("ERROR cloning samples")
	}
	json.Unmarshal(raw, &applications)
	return
}

//
// Create a TagCategory.
func Create(t *testing.T, tagCategory *api.TagCategory) {
	err := Client.Post(api.TagCategoriesRoot, &tagCategory)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error()) // Fatal here, Error for standard test failure or failed assertion.
	}
}

//
// Delete the TagCategory.
func Delete(t *testing.T, tagCategory *api.TagCategory) {
	err := Client.Delete(fmt.Sprintf("%s/%d", api.TagCategoriesRoot, tagCategory.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
