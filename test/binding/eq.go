package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	"github.com/konveyor/tackle2-hub/test/cmp/sort"
)

func init() {
	sort.Add(
		sort.ById,
		[]api.Ref{},
		[]*api.Ref{})
	cmp.IgnoredPaths = []string{
		"CreateUser",
		"UpdateUser",
		"CreateTime",
		"UpdateTime",
		"Profiles.UpdateTime",
		"Profiles.UpdateUser",
		"Profiles.CreateUser",
		"Insights.UpdateTime",
		"Insights.UpdateUser",
		"Insights.CreateUser",
		"Insights.Incidents.UpdateTime",
		"Insights.Incidents.UpdateUser",
		"Insights.Incidents.CreateUser",
		"Dependencies.UpdateTime",
		"Dependencies.UpdateUser",
		"Dependencies.CreateUser",
	}
}
