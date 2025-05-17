package manifest

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Base = api.Generator{
		Kind: "base",
		Name: "Test",
		Repository: &api.Repository{
			URL: "https://github.com/konveyor/tackle2-hub",
		},
		Parameters: api.Map{
			"p1": "v1",
			"p2": "v2",
		},
		Values: api.Map{
			"p1": "v1",
			"p2": "v2",
		},
	}
)
