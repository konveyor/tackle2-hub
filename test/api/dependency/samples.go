package dependency

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Struct to hold dependency sample data.
type DependencySample struct {
	ApplicationFrom api.Application
	ApplicationTo   api.Application
}

type ReverseDependencySample struct {
	Application1 api.Application
	Application2 api.Application
	Application3 api.Application
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

var ReverseSamples = []ReverseDependencySample{
	{
		Application1: api.Application{
			Name:        "First",
			Description: "First application",
		},
		Application2: api.Application{
			Name:        "Second",
			Description: "Second application",
		},
		Application3: api.Application{
			Name:        "Third",
			Description: "Third application",
		},
	},
}
