package jobfunction

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
// Set of valid resources for tests and reuse.
var Samples = []*api.JobFunction{
	{
		Name: "Engineer",
	},
	{
		Name: "Manager",
	},
}

//
// Creates a copy of Samples for given test.
func CloneSamples() (samples []*api.JobFunction) {
	raw, err := json.Marshal(Samples)
	if err != nil {
		panic("ERROR cloning samples")
	}
	json.Unmarshal(raw, &samples)

	return
}

//
// Create a Tag.
func Create(t *testing.T, r *api.JobFunction) {
	err := Client.Post(api.JobFunctionsRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error())
	}
}

//
// Delete the Tag.
func Delete(t *testing.T, r *api.JobFunction) {
	err := Client.Delete(fmt.Sprintf("%s/%d", api.JobFunctionsRoot, r.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
