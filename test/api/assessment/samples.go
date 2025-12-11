package assessment

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/api/application"
	"github.com/konveyor/tackle2-hub/test/api/questionnaire"
)

// Set of valid resources for tests and reuse.
var (
	ApplicationAssessment1 = api.Assessment{
		// Ref resource are created by the test.
		Application: &api.Ref{
			Name: application.Minimal.Name,
		},
		Questionnaire: api.Ref{
			Name: questionnaire.Questionnaire1.Name,
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
	Samples = []api.Assessment{ApplicationAssessment1}
)
