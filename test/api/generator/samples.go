package manifest

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Base = api2.Generator{
		Kind:        "base",
		Name:        "Test",
		Description: "This is a test",
		Repository: &api2.Repository{
			URL: "https://github.com/konveyor/tackle2-hub",
		},
		Params: api2.Map{
			"p1": "v1",
			"p2": "v2",
		},
		Values: api2.Map{
			"p1": "v1",
			"p2": "v2",
		},
	}
)
