package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestAssessment(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a questionnaire for the assessment to reference
	questionnaire := &api.Questionnaire{
		Name:        "Test Assessment Questionnaire",
		Description: "Questionnaire for assessment test",
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
	err := client.Questionnaire.Create(questionnaire)
	g.Expect(err).To(BeNil())
	g.Expect(questionnaire.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Questionnaire.Delete(questionnaire.ID)
	})

	// Create an application for the assessment to reference
	application := &api.Application{
		Name:        "Test Assessment Application",
		Description: "Application for assessment test",
	}
	err = client.Application.Create(application)
	g.Expect(err).To(BeNil())
	g.Expect(application.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

	// Define the assessment to create
	assessment := &api.Assessment{
		Application: &api.Ref{
			ID:   application.ID,
			Name: application.Name,
		},
		Questionnaire: api.Ref{
			ID:   questionnaire.ID,
			Name: questionnaire.Name,
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

	// CREATE: Create the assessment via Application
	err = client.Application.Select(application.ID).Assessment.Create(assessment)
	g.Expect(err).To(BeNil())
	g.Expect(assessment.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment.ID)
	})

	// GET: List assessments
	list, err := client.Assessment.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(assessment, list[0], "Required")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the assessment and verify it matches
	retrieved, err := client.Assessment.Get(assessment.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(assessment, retrieved, "Required")
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the assessment - change answer selection
	assessment.Sections[0].Questions[0].Answers[2].Selected = false // Deselect Blue
	assessment.Sections[0].Questions[0].Answers[1].Selected = true  // Select Green

	err = client.Assessment.Update(assessment)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Assessment.Get(assessment.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	g.Expect(updated.Sections[0].Questions[0].Answers[2].Selected).To(BeFalse()) // Blue not selected
	g.Expect(updated.Sections[0].Questions[0].Answers[1].Selected).To(BeTrue())  // Green selected

	// DELETE: Remove the assessment
	err = client.Assessment.Delete(assessment.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Assessment.Get(assessment.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
