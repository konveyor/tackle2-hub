package ticket

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

var Samples = []api.Ticket{
	{
		Kind:   "10001",
		Parent: "10000",
		Application: api.Ref{
			Name: "Sample Application1",
		},
		Tracker: api.Ref{
			Name: "Sample Ticket-Tracker",
		},
	},
}
