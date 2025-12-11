package review

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

var Samples = []api.Review{
	{
		BusinessCriticality: 1,
		EffortEstimate:      "min",
		ProposedAction:      "run",
		WorkPriority:        1,
		Comments:            "nil",
		Application: &api.Ref{
			Name: "Sample Review 1",
		},
	},
	{
		BusinessCriticality: 2,
		EffortEstimate:      "max",
		ProposedAction:      "stop",
		WorkPriority:        2,
		Comments:            "nil",
		Application: &api.Ref{
			Name: "Sample Review 2",
		},
	},
}
