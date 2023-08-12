package tracker

import (
	"github.com/konveyor/tackle2-hub/api"
)

var Samples = []api.Tracker{
	{
		Name:    "Sample tracker",
		URL:     "https://konveyor.io/test/api/tracker",
		Kind:    "jira-onprem",
		Message: "Description of tracker",
		Identity: api.Ref{
			ID:   1,
			Name: "Sample Tracker Identity",
		},
	},
	{
		Name:    "Sample tracker1",
		URL:     "https://konveyor.io/test/api/tracker1",
		Kind:    "jira-cloud",
		Message: "Description of tracker1",
		Identity: api.Ref{
			ID:   2,
			Name: "Sample Tracker Identity1",
		},
	},
}
