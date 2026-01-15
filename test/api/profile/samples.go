package profile

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

var (
	Base = api.AnalysisProfile{
		Name:        "Test",
		Description: "This is a test analysis profile",
		Mode:        api.ApMode{WithDeps: true},
		Scope: api.ApScope{
			WithKnownLibs: true,
			Packages: api.InExList{
				Included: []string{"pA", "pB"},
				Excluded: []string{"pC", "pD"},
			},
		},
		Rules: api.ApRules{
			Targets: []api.ApTargetRef{
				{ID: 2, Name: "Containerization"},
				{ID: 6, Name: "OpenJDK", Selection: "konveyor.io/target=openjdk17"},
			},
			Labels: api.InExList{
				Included: []string{"rA", "rB"},
				Excluded: []string{"rC", "rD"},
			},
			Repository: &api.Repository{
				URL:  "https://github.com/konveyor/rulesets.git",
				Path: "default/generated/camel3",
			},
		}}
)
