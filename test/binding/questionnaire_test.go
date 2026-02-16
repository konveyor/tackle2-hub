package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestQuestionnaire(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the questionnaire to create
	questionnaire := &api.Questionnaire{
		Name:        "Test Questionnaire",
		Description: "Questionnaire test sample",
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
						},
					},
				},
			},
		},
	}

	// Get seeded.
	seeded, err := client.Questionnaire.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the questionnaire
	err = client.Questionnaire.Create(questionnaire)
	g.Expect(err).To(BeNil())
	g.Expect(questionnaire.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Questionnaire.Delete(questionnaire.ID)
	})

	// GET: List questionnaires
	list, err := client.Questionnaire.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(questionnaire, list[len(seeded)])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the questionnaire and verify it matches
	retrieved, err := client.Questionnaire.Get(questionnaire.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(questionnaire, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the questionnaire
	questionnaire.Name = "Updated Test Questionnaire"
	questionnaire.Description = "Updated questionnaire description"
	questionnaire.Required = false

	err = client.Questionnaire.Update(questionnaire)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Questionnaire.Get(questionnaire.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(questionnaire, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the questionnaire
	err = client.Questionnaire.Delete(questionnaire.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Questionnaire.Get(questionnaire.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestQuestionnaireValidation tests all validation rules in api.Questionnaire.Validate()
func TestQuestionnaireValidation(t *testing.T) {
	g := NewGomegaWithT(t)

	// Valid questionnaire template based on tackle2-ui example
	validQuestionnaire := api.Questionnaire{
		Name:        "Cloud Readiness Questionnaire",
		Description: "Assess cloud readiness of applications",
		Required:    true,
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Application Technologies",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "What is the main technology in your application?",
						Explanation: "Identify the main framework or technology used.",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Quarkus",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "Spring Boot",
								Risk:  "green",
							},
							{
								Order: 3,
								Text:  "Legacy Monolithic Application",
								Risk:  "red",
							},
						},
					},
					{
						Order:       2,
						Text:        "Does your application use a microservices architecture?",
						Explanation: "Assess if the application is built using microservices.",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Yes",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "No",
								Risk:  "yellow",
							},
							{
								Order: 3,
								Text:  "Unknown",
								Risk:  "unknown",
							},
						},
					},
				},
			},
			{
				Order: 2,
				Name:  "Data Storage",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "Is your application's data storage cloud-optimized?",
						Explanation: "Evaluate if the data storage solution is optimized for cloud.",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Cloud-Native Storage Solution",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "Traditional On-Premises Storage",
								Risk:  "red",
							},
							{
								Order: 3,
								Text:  "Hybrid Storage Approach",
								Risk:  "yellow",
							},
						},
					},
				},
			},
		},
		Thresholds: api.Thresholds{
			Red:     1,
			Yellow:  30,
			Unknown: 15,
		},
		RiskMessages: api.RiskMessages{
			Red:     "Requires deep changes in architecture or lifecycle",
			Yellow:  "Cloud friendly but needs minor changes",
			Green:   "Cloud Native",
			Unknown: "More information needed",
		},
	}

	tests := []struct {
		name          string
		questionnaire api.Questionnaire
		wantError     bool
		errorContains string
	}{
		{
			name:          "Valid questionnaire",
			questionnaire: validQuestionnaire,
			wantError:     false,
		},
		{
			name: "Duplicate section order",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "Question 1",
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "green"},
								},
							},
						},
					},
					{
						Order: 1, // Duplicate order
						Name:  "Section 2",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "Question 1",
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "green"},
								},
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "duplicate section order 1 found",
		},
		{
			name: "Section with no questions",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order:     1,
						Name:      "Empty Section",
						Questions: []api.Question{}, // No questions
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "Questions", // Binding validation catches this first
		},
		{
			name: "Duplicate question order within section",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "Question 1",
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "green"},
								},
							},
							{
								Order: 1, // Duplicate order
								Text:  "Question 2",
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "green"},
								},
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "duplicate question order 1 found",
		},
		{
			name: "Question with empty text",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "", // Empty text
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "green"},
								},
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "must have text",
		},
		{
			name: "Question with no answers",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order:   1,
								Text:    "Question with no answers",
								Answers: []api.Answer{}, // No answers
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "Answers", // Binding validation catches this first
		},
		{
			name: "Duplicate answer order within question",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "Question 1",
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "green"},
									{Order: 1, Text: "Answer 2", Risk: "yellow"}, // Duplicate order
								},
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "duplicate answer order 1 found",
		},
		{
			name: "Answer with empty text",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "Question 1",
								Answers: []api.Answer{
									{Order: 1, Text: "", Risk: "green"}, // Empty text
								},
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "must have text",
		},
		{
			name: "Answer with invalid risk level",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Sections = []api.Section{
					{
						Order: 1,
						Name:  "Section 1",
						Questions: []api.Question{
							{
								Order: 1,
								Text:  "Question 1",
								Answers: []api.Answer{
									{Order: 1, Text: "Answer 1", Risk: "invalid"}, // Invalid risk
								},
							},
						},
					},
				}
				return q
			}(),
			wantError:     true,
			errorContains: "Risk", // Binding validation catches this first with 'oneof' tag
		},
		{
			name: "All thresholds are zero",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.Thresholds = api.Thresholds{
					Red:     0,
					Yellow:  0,
					Unknown: 0,
				}
				return q
			}(),
			wantError:     true,
			errorContains: "at least one threshold",
		},
		{
			name: "Missing red risk message",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.RiskMessages = api.RiskMessages{
					Red:     "", // Empty
					Yellow:  "Yellow message",
					Green:   "Green message",
					Unknown: "Unknown message",
				}
				return q
			}(),
			wantError:     true,
			errorContains: "all risk messages",
		},
		{
			name: "Missing yellow risk message",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.RiskMessages = api.RiskMessages{
					Red:     "Red message",
					Yellow:  "", // Empty
					Green:   "Green message",
					Unknown: "Unknown message",
				}
				return q
			}(),
			wantError:     true,
			errorContains: "all risk messages",
		},
		{
			name: "Missing green risk message",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.RiskMessages = api.RiskMessages{
					Red:     "Red message",
					Yellow:  "Yellow message",
					Green:   "", // Empty
					Unknown: "Unknown message",
				}
				return q
			}(),
			wantError:     true,
			errorContains: "all risk messages",
		},
		{
			name: "Missing unknown risk message",
			questionnaire: func() api.Questionnaire {
				q := validQuestionnaire
				q.RiskMessages = api.RiskMessages{
					Red:     "Red message",
					Yellow:  "Yellow message",
					Green:   "Green message",
					Unknown: "", // Empty
				}
				return q
			}(),
			wantError:     true,
			errorContains: "all risk messages",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Questionnaire.Create(&tt.questionnaire)
			if err == nil {
				id := tt.questionnaire.ID
				if id != 0 {
					t.Cleanup(func() {
						_ = client.Questionnaire.Delete(id)
					})
				}
			}
			if tt.wantError {
				g.Expect(err).ToNot(BeNil(), "Expected an error.")
				if tt.errorContains != "" {
					g.Expect(err.Error()).
						To(ContainSubstring(tt.errorContains), "Error message mismatch.")
				}
			} else {
				g.Expect(err).To(BeNil(), "Error not expected.")
			}
		})
	}
}
