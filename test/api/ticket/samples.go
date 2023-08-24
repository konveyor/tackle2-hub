package ticket

import (
	"github.com/konveyor/tackle2-hub/api"
	TrackerSamples "github.com/konveyor/tackle2-hub/test/api/tracker"
)

var Samples = []api.Ticket{
	{
		Kind:      "jira-onprem",
		Reference: "Sample Reference for Ticket",
		Link:      "www.konveyor.io/ticket",
		Parent:    "Sample Parent Ticket",
		Application: api.Ref{
			ID:   1,
			Name: "Sample Application1",
		},
		Tracker: api.Ref{
			ID:   1,
			Name: TrackerSamples.Samples[0].Name,
		},
	},
}
