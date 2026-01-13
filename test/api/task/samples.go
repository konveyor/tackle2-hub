package task

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Windup = api.Task{
		Name:  "Test",
		Addon: "analyzer",
		Data:  api.Map{},
	}
	Samples = []api.Task{Windup}
)
