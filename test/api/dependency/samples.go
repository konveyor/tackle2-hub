package dependency

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Struct to hold dependency sample data.
type DependencySample struct {
	ApplicationFrom api.Application
	ApplicationTo   api.Application
}

var Samples = []DependencySample{
	{
		ApplicationFrom: api.Application{
			Name:        "Gateway",
			Description: "Gateway application",
		},
		ApplicationTo: api.Application{
			Name:        "Inventory",
			Description: "Inventory application",
		},
	},
}
