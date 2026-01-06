package assessment

import (
	api2 "github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/internal/test/api/application"
	"github.com/konveyor/tackle2-hub/internal/test/api/questionnaire"
)

// Set of valid resources for tests and reuse.
var (
	ApplicationAssessment1 = api2.Assessment{
		// Ref resource are created by the test.
		Application: &api2.Ref{
			Name: application.Minimal.Name,
		},
		Questionnaire: api2.Ref{
			Name: questionnaire.Questionnaire1.Name,
		},
		Sections: []api2.Section{
			{
				Order: 1,
				Name:  "Section 1",
				Questions: []api2.Question{
					{
						Order:       1,
						Text:        "What is your favorite color?",
						Explanation: "Please tell us your favorite color.",
						Answers: []api2.Answer{
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
	Samples = []api2.Assessment{ApplicationAssessment1}
)
