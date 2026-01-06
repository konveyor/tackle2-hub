package tag

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	TestLinux = api2.Tag{
		Name: "Test Linux",
		Category: api2.Ref{
			ID: 1, // Category from seeds.
		},
	}
	TestRHEL = api2.Tag{
		Name: "Test RHEL",
		Category: api2.Ref{
			ID: 2, // Category from seeds.
		},
	}
	Samples = []api2.Tag{TestLinux, TestRHEL}
)
