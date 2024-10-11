package questionnaire

import (
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/model"
)

// Set of valid resources for tests and reuse.
var (
	Questionnaire1 = api.Questionnaire{
		Name:         "Questionnaire1",
		Description:  "Questionnaire minimal sample 1",
		Required:     true,
		Thresholds:   api.Thresholds{},
		RiskMessages: api.RiskMessages{},
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Section 1",
				Questions: []model.Question{
					{
						Order:       1,
						Text:        "What is your favorite color?",
						Explanation: "Please tell us your favorite color.",
						Answers: []model.Answer{
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
