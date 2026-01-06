package ticket

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

var Samples = []api2.Ticket{
	{
		Kind:   "10001",
		Parent: "10000",
		Application: api2.Ref{
			Name: "Sample Application1",
		},
		Tracker: api2.Ref{
			Name: "Sample Ticket-Tracker",
		},
	},
}
