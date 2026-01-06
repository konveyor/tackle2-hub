package review

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

var Samples = []api2.Review{
	{
		BusinessCriticality: 1,
		EffortEstimate:      "min",
		ProposedAction:      "run",
		WorkPriority:        1,
		Comments:            "nil",
		Application: &api2.Ref{
			Name: "Sample Review 1",
		},
	},
	{
		BusinessCriticality: 2,
		EffortEstimate:      "max",
		ProposedAction:      "stop",
		WorkPriority:        2,
		Comments:            "nil",
		Application: &api2.Ref{
			Name: "Sample Review 2",
		},
	},
}
