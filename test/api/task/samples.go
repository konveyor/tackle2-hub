package task

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Windup = api.Task{
		Name:  "Test task",
		Addon: "Test",
		Data:  "{}",
	}
	Samples = []api.Task{Windup}
)
