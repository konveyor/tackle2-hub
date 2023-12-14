package assessment

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/assessment"
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
			ID:   1,
			Name: questionnaire.Questionnaire1.Name,
		},
		Sections: []assessment.Section{
			{
				Order: uint(1),
				Name:  "Section 1",
				Questions: []assessment.Question{
					{
						Order:       uint(1),
						Text:        "What is your favorite color?",
						Explanation: "Please tell us your favorite color.",
						Answers: []assessment.Answer{
							{
								Order: uint(1),
								Text:  "Red",
								Risk:  "red",
							},
							{
								Order: uint(2),
								Text:  "Green",
								Risk:  "green",
							},
							{
								Order:    uint(3),
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
