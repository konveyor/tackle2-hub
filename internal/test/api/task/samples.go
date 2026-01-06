package task

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Windup = api2.Task{
		Name:  "Test",
		Addon: "analyzer",
		Data:  api2.Map{},
	}
	Samples = []api2.Task{Windup}
)
