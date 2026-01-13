package application

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid Application resources for tests and reuse.
// Important: initialize test application from this samples, not use it directly to not affect other tests.
var (
	Minimal = api2.Application{
		Name: "Minimal application",
	}
	PathfinderGit = api2.Application{
		Name:        "Pathfinder",
		Description: "Tackle Pathfinder application.",
		Repository: &api2.Repository{
			Kind:   "git",
			URL:    "https://github.com/konveyor/tackle-pathfinder.git",
			Branch: "1.2.0",
		},
	}
	Samples = []api2.Application{Minimal, PathfinderGit}
)
