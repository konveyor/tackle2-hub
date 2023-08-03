package target

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Minimal = api.Target{
		Name: "Minimal no ruleset",
	}

	Hazelcast = api.Target{
		Name: "Hazelcast",
		Image: api.Ref{
			ID: 1,
		},
		Description: "Hazelcast Java distributed session store.",
		RuleSet: &api.RuleSet{
			Rules: []api.Rule{
				{
					File: &api.Ref{
						Name: "./data/rules.yaml",
					},
				},
			},
		},
	}
	Samples = []api.Target{Minimal, Hazelcast}
)
