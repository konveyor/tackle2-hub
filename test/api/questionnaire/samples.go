package questionnaire

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/assessment"
)

// Set of valid resources for tests and reuse.
var (
	Questionnaire1 = api.Questionnaire{
		Name:         "Questionnaire1",
		Description:  "Questionnaire minimal sample 1",
		Required:     true,
		Thresholds:   assessment.Thresholds{},
		RiskMessages: assessment.RiskMessages{},
		Sections: []assessment.Section{
			{
				Order: 1,
				Name:  "Section 1",
				Questions: []assessment.Question{
					{
						Order:       1,
						Text:        "What is your favorite color?",
						Explanation: "Please tell us your favorite color.",
						Answers: []assessment.Answer{
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
