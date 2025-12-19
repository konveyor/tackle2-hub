package questionnaire

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
)

// TestQuestionnaireValidation tests all validation rules in api.Questionnaire.Validate()
func TestQuestionnaireValidation(t *testing.T) {
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
			err := Questionnaire.Create(&tt.questionnaire)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.errorContains)
					// Clean up if it was created
					if tt.questionnaire.ID != 0 {
						_ = Questionnaire.Delete(tt.questionnaire.ID)
					}
					return
				}

				// Check if error contains expected string
				if tt.errorContains != "" {
					errStr := err.Error()
					if !contains(errStr, tt.errorContains) {
						t.Errorf("Expected error containing '%s', but got: %s", tt.errorContains, errStr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %s", err.Error())
					return
				}

				// Clean up valid questionnaire
				if tt.questionnaire.ID != 0 {
					err = Questionnaire.Delete(tt.questionnaire.ID)
					if err != nil {
						t.Errorf("Failed to delete questionnaire: %s", err.Error())
					}
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
