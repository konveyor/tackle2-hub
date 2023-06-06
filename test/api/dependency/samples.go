package dependency

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
type TestCase struct {
	Name            string
	ApplicationFrom api.Application
	ApplicationTo   api.Application
}

var (
	dependency = api.Dependency{
		To: api.Ref{
			Name: "Gateway",
		},
		From: api.Ref{
			Name: "Inventory",
		},
	}

	Gateway = api.Application{
		Name:        "Gateway",
		Description: "Gateway application",
	}
	Inventory = api.Application{
		Name:        "Inventory",
		Description: "Inventory application",
	}

	testCase = TestCase{
		Name:            "Test",
		ApplicationFrom: Gateway,
		ApplicationTo:   Inventory,
	}
	Samples = []api.Dependency{}
)

func init() {
	tc := []TestCase{testCase}
	for _, t := range tc {
		dependency := api.Dependency{
			To: api.Ref{
				Name: t.ApplicationTo.Name,
			},
			From: api.Ref{
				Name: t.ApplicationFrom.Name,
			},
		}
		Samples = []api.Dependency{dependency}
	}
}
