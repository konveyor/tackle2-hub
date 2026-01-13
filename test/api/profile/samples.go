package profile

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

var (
	Base = api2.AnalysisProfile{
		Name:        "Test",
		Description: "This is a test analysis profile",
		Mode:        api2.ApMode{WithDeps: true},
		Scope: api2.ApScope{
			WithKnownLibs: true,
			Packages: api2.InExList{
				Included: []string{"pA", "pB"},
				Excluded: []string{"pC", "pD"},
			},
		},
		Rules: api2.ApRules{
			Targets: []api2.Ref{
				{ID: 2, Name: "Containerization"},
			},
			Labels: api2.InExList{
				Included: []string{"rA", "rB"},
				Excluded: []string{"rC", "rD"},
			},
			Repository: &api2.Repository{
				URL:  "https://github.com/konveyor/rulesets.git",
				Path: "default/generated/camel3",
			},
		}}
)
