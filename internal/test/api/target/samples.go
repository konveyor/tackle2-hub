package target

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Minimal = api2.Target{
		Name: "Minimal no ruleset",
		Image: api2.Ref{
			Name: "./data/image.svg",
		},
	}

	Hazelcast = api2.Target{
		Name: "Hazelcast",
		Image: api2.Ref{
			Name: "./data/image.svg",
		},
		Description: "Hazelcast Java distributed session store.",
		RuleSet: &api2.RuleSet{
			Rules: []api2.Rule{
				{
					File: &api2.Ref{
						Name: "./data/rules.yaml",
					},
				},
				{
					File: &api2.Ref{
						Name: "./data/rules.yaml",
					},
				},
			},
		},
	}
	Samples = []api2.Target{Minimal, Hazelcast}
)
