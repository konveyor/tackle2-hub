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

	// CREATE: Create the questionnaire
	err := client.Questionnaire.Create(questionnaire)
	g.Expect(err).To(BeNil())
	g.Expect(questionnaire.ID).NotTo(BeZero())

	defer func() {
		_ = client.Questionnaire.Delete(questionnaire.ID)
	}()

	// GET: Retrieve the questionnaire and verify it matches
	retrieved, err := client.Questionnaire.Get(questionnaire.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(questionnaire, retrieved)
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
