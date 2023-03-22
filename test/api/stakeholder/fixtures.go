package stakeholder

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
var Samples = []*api.Stakeholder{
	{
		Name:  "Alice",
		Email: "alice@acme.local",
	},
	{
		Name:  "Bob",
		Email: "bob@acme-supplier.local",
	},
}

//
// Creates a copy of Samples for a test.
func CloneSamples() (samples []*api.Stakeholder) {
	raw, err := json.Marshal(Samples)
	if err != nil {
		panic("ERROR cloning samples")
	}
	json.Unmarshal(raw, &samples)
	return
}

//
// Create.
func Create(t *testing.T, r *api.Stakeholder) {
	err := Client.Post(api.StakeholdersRoot, &r)
	if err != nil {
		t.Fatalf("Create fatal error: %v", err.Error())
	}
}

//
// Delete.
func Delete(t *testing.T, r *api.Stakeholder) {
	err := Client.Delete(fmt.Sprintf("%s/%d", api.StakeholdersRoot, r.ID))
	if err != nil {
		t.Fatalf("Delete fatal error: %v", err.Error())
	}
}
