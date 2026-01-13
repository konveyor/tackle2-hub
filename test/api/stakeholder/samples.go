package stakeholder

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Alice = api.Stakeholder{
		Name:  "Alice",
		Email: "alice@acme.local",
	}
	Bob = api.Stakeholder{
		Name:  "Bob",
		Email: "bob@acme-supplier.local",
	}
	Samples = []api.Stakeholder{Alice, Bob}
)
