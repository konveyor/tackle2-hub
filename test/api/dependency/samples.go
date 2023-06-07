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

var Samples = []api.Dependency{}

var (
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

	dependency = api.Dependency{
		To: api.Ref{
			Name: testCase.ApplicationTo.Name,
		},
		From: api.Ref{
			Name: testCase.ApplicationFrom.Name,
		},
	}
)

func init() {
	Samples = []api.Dependency{dependency}
}
