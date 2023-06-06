package dependency

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
type TestCase struct {
	Name         string
	Dependency   api.Dependency
	Application1 api.Application
	Application2 api.Application
}

var (
	dependency = api.Dependency{
		To: api.Ref{
			ID:   uint(1),
			Name: "Alice",
		},
		From: api.Ref{
			ID:   uint(2),
			Name: "Bob",
		},
	}

	aliceApplication = api.Application{
		Name:        "Alice",
		Description: "alice's application",
	}
	bobApplication = api.Application{
		Name:        "Bob",
		Description: "bob's application",
	}

	testCase = TestCase{
		Name:         "Test",
		Dependency:   dependency,
		Application1: aliceApplication,
		Application2: bobApplication,
	}
	tc      = []TestCase{testCase}
	Samples = []api.Dependency{tc}
)
