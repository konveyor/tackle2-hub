package tag

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	TestLinux = api.Tag{
		Name: "Test Linux",
		Category: api.Ref{
			ID: 1, // Category from seeds.
		},
	}
	TestRHEL = api.Tag{
		Name: "Test RHEL",
		Category: api.Ref{
			ID: 2, // Category from seeds.
		},
	}
	Samples = []api.Tag{TestLinux, TestRHEL}
)
