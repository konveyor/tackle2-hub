package jobfunction

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Engineer = api.JobFunction{
		Name: "Engineer",
	}
	Manager = api.JobFunction{
		Name: "Manager",
	}
	Samples = []api.JobFunction{Engineer, Manager}
)
