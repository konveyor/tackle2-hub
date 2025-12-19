package manifest

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Base = api.Platform{
		Kind: "base",
		Name: "Test",
		URL:  "http://localhost",
	}
)
