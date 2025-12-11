package analysisprofile

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
			Targets: []api.Ref{
				{ID: 2, Name: "Containerization"},
			},
			Labels: api.InExList{
				Included: []string{"rA", "rB"},
				Excluded: []string{"rC", "rD"},
			},
			Files: []api.Ref{
				{ID: 400},
			},
			Repository: &api.Repository{
				URL: "https://github.com/konveyor/testapp",
			},
		}}
)
