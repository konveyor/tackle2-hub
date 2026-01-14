package businessservice

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Marketing = api.BusinessService{
		Name:        "Marketing",
		Description: "Marketing dept service.",
	}
	Sales = api.BusinessService{
		Name:        "Sales",
		Description: "Sales support service.",
	}
	Samples = []api.BusinessService{Marketing, Sales}
)
