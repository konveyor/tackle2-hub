package analysisprofile

import (
	"github.com/konveyor/tackle2-hub/api"
)

var (
	Base = api.AnalysisProfile{}
)

func init() {
	Base.Name = "base"
	Base.Description = "Base analysis profiling test."
	Base.Mode.WithDeps = true
	Base.Scope.WithKnownLibs = true
	Base.Scope.Packages.Included = []string{"pA", "pB"}
	Base.Scope.Packages.Excluded = []string{"pC", "pD"}
	Base.Rules.Targets = []api.Ref{
		{ID: 2, Name: "Containerization"},
	}
	Base.Rules.Labels.Included = []string{"A", "B"}
	Base.Rules.Labels.Excluded = []string{"C", "D"}
	Base.Rules.Files = []api.Ref{
		{ID: 400},
	}
	Base.Rules.Repository = &api.Repository{
		URL: "https://github.com/konveyor/testapp",
	}
}
