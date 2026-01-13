package ruleset

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Minimal = api2.RuleSet{
		Name:  "Minimal no rules",
		Rules: []api2.Rule{},
	}

	Hazelcast = api2.RuleSet{
		Name:        "Hazelcast",
		Description: "Hazelcast Java distributed session store ruleset.",
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
	}
	Samples = []api2.RuleSet{Minimal, Hazelcast}
)
