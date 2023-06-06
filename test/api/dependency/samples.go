package dependency

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	firstDependency = api.Dependency{
		To: api.Ref{
			ID:   uint(1473),
			Name: "Alice",
		},
		From: api.Ref{
			ID:   uint(2),
			Name: "Bob",
		},
	}

	secondDependency = api.Dependency{
		To: api.Ref{
			ID:   uint(2123),
			Name: "Bob",
		},
		From: api.Ref{
			ID:   uint(1),
			Name: "Alice",
		},
	}

	Samples = []api.Dependency{firstDependency, secondDependency}
)
