package tag

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/api/tagcategory"
)

var (
	// Setup Hub API client
	Client = client.Client
)

//
// Set of valid resources for tests and reuse.
var Samples = []*api.Tag{
	{
		Name:     "Test Linux",
		Category: api.Ref{
			//ID:   99901, // or not hardcode ID here
			//Name: "Test OS categories",
		},
	},
	{
		Name:     "Test RHEL",
		Category: api.Ref{
			//ID:   99902, // or not hardcode ID here
			//Name: "Test Linux distros",
		},
	},
}

//
// Creates a copy of Samples for given test (copy is there to avoid tests inflence each other using the same object ref).
func CloneSamples() (samples []*api.Tag) {
	raw, err := json.Marshal(Samples)
	if err != nil {
		fmt.Print("ERROR cloning samples")
	}
	json.Unmarshal(raw, &samples)

	return
}

//
// Creates category for Tag test from tagcategory package samples.
func prepareCategory(t *testing.T, tags []*api.Tag) (category *api.TagCategory) {
	category = tagcategory.CloneSamples()[0]
	tagcategory.Create(t, category)

	// Assign category to Tag.
	for _, tag := range tags {
		tag.Category = api.Ref{ID: category.ID}
	}
	return
}

//
// Create a Tag.
func Create(t *testing.T, r *api.Tag) {
	err := Client.Post(api.TagsRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error()) // Fatal here, Error for standard test failure or failed assertion.
	}
}

//
// Delete the Tag.
func Delete(t *testing.T, r *api.Tag) {
	err := Client.Delete(fmt.Sprintf("%s/%d", api.TagsRoot, r.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
