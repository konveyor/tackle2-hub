package ticket

import (
	"github.com/konveyor/tackle2-hub/api"
)

var Samples = []api.Ticket{
	{
		Kind:      "Sample Tracker",
		Reference: "Sample Reference",
		Link:      "www.konveyor.io/ticket",
		Parent:    "Sample Parent",
		Application: api.Ref{
			ID:   1,
			Name: "Sample Application1",
		},
		Tracker: api.Ref{
			ID:   1,
			Name: "Sample Tracker1",
		},
	},
}
