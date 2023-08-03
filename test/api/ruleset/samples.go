package ruleset

import (
	"github.com/konveyor/tackle2-hub/api"
)

// Set of valid resources for tests and reuse.
var (
	Minimal = api.RuleSet{
		Name: "Minimal no rules",
		Rules: []api.Rule{},
	}

	Hazelcast = api.RuleSet{
		Name:        "Hazelcast",
		Description: "Hazelcast Java distributed session store ruleset.",
		Rules: []api.Rule{
			{
				File: &api.Ref{
					Name: "./data/hz.windup.xml",
				},
			},
		},
	}
	Samples = []api.RuleSet{Minimal, Hazelcast}
)
