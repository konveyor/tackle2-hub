package questionnaire

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	Questionnaire1 = api.Questionnaire{
		Name:        "Questionnaire1",
		Description: "Questionnaire minimal sample 1",
		Required:    true,
		Thresholds: api.Thresholds{
			Red:     30,
			Yellow:  20,
			Unknown: 10,
		},
		RiskMessages: api.RiskMessages{
			Red:     "Application requires deep changes in code or infrastructure.",
			Yellow:  "Application requires some changes in code or configuration.",
			Green:   "Application is ready for modernization.",
			Unknown: "More information needed.",
		},
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Section 1",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "What is your favorite color?",
						Explanation: "Please tell us your favorite color.",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Red",
								Risk:  "red",
							},
							{
								Order: 2,
								Text:  "Green",
								Risk:  "green",
							},
							{
								Order:    3,
								Text:     "Blue",
								Risk:     "yellow",
								Selected: true,
							},
						},
					},
				},
			},
		},
	}
	Samples = []api.Questionnaire{Questionnaire1}
)
