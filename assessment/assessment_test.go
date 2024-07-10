package assessment

import (
	"testing"

	"github.com/konveyor/tackle2-hub/model"
	"github.com/onsi/gomega"
)

func TestPrepareSections(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	sections := []Section{
		{
			Questions: []Question{
				{
					Text: "Default",
					Answers: []Answer{
						{
							Text: "Answer1",
						},
						{
							Text: "Answer2",
						},
					},
				},
				{
					Text: "Should Include",
					IncludeFor: []CategorizedTag{
						{Category: "Category", Tag: "Tag"},
					},
					Answers: []Answer{
						{
							Text: "Answer1",
						},
						{
							Text: "Answer2",
						},
					},
				},
				{
					Text: "Should Exclude",
					ExcludeFor: []CategorizedTag{
						{Category: "Category", Tag: "Tag"},
					},
					Answers: []Answer{
						{
							Text: "Answer1",
						},
						{
							Text: "Answer2",
						},
					},
				},
				{
					Text: "AutoAnswer",
					Answers: []Answer{
						{
							Text: "Answer1",
							AutoAnswerFor: []CategorizedTag{
								{Category: "Category", Tag: "Tag"},
							},
						},
						{
							Text: "Answer2",
						},
					},
				},
			},
		},
	}
	tagResolver := TagResolver{
		cache: map[string]map[string]*model.Tag{
			"Category": {"Tag": {Model: model.Model{ID: 1}}},
		},
	}
	tags := NewSet()
	tags.Add(1)

	preparedSections := prepareSections(&tagResolver, tags, sections)
	questions := preparedSections[0].Questions

	g.Expect(len(questions)).To(gomega.Equal(3))
	g.Expect(questions[0].Text).To(gomega.Equal("Default"))
	g.Expect(questions[0].Answered()).To(gomega.BeFalse())
	g.Expect(questions[1].Text).To(gomega.Equal("Should Include"))
	g.Expect(questions[1].Answered()).To(gomega.BeFalse())
	g.Expect(questions[2].Text).To(gomega.Equal("AutoAnswer"))
	g.Expect(questions[2].Answered()).To(gomega.BeTrue())
	g.Expect(questions[2].Answers[0].Text).To(gomega.Equal("Answer1"))
	g.Expect(questions[2].Answers[0].AutoAnswered).To(gomega.BeTrue())
	g.Expect(questions[2].Answers[0].Selected).To(gomega.BeTrue())
}

func TestAssessmentStarted(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	assessment := Assessment{
		Sections: []Section{
			{
				Questions: []Question{
					{
						Text: "S1Q1",
						Answers: []Answer{
							{
								Text:     "A1",
								Selected: true,
							},
							{
								Text: "A2",
							},
						},
					},
					{
						Text: "S1Q2",
						Answers: []Answer{
							{
								Text: "A1",
							},
							{
								Text: "A2",
							},
						},
					},
				},
			},
			{
				Questions: []Question{
					{
						Text: "S2Q1",
						Answers: []Answer{
							{
								Text: "A1",
							},
							{
								Text: "A2",
							},
						},
					},
				},
			},
		},
	}
	g.Expect(assessment.Started()).To(gomega.BeTrue())
	g.Expect(assessment.Status()).To(gomega.Equal(StatusStarted))
	assessment.Sections[0].Questions[0].Answers[0].AutoAnswered = true
	g.Expect(assessment.Started()).To(gomega.BeFalse())
	g.Expect(assessment.Status()).To(gomega.Equal(StatusEmpty))
}

func TestAssessmentComplete(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	assessment := Assessment{
		Sections: []Section{
			{
				Questions: []Question{
					{
						Text: "S1Q1",
						Answers: []Answer{
							{
								Text: "A1",
							},
							{
								Text: "A2",
							},
						},
					},
					{
						Text: "S1Q2",
						Answers: []Answer{
							{
								Text:     "A1",
								Selected: true,
							},
							{
								Text: "A2",
							},
						},
					},
				},
			},
			{
				Questions: []Question{
					{
						Text: "S2Q1",
						Answers: []Answer{
							{
								Text: "A1",
							},
							{
								Text:         "A2",
								Selected:     true,
								AutoAnswered: true,
							},
						},
					},
				},
			},
		},
	}
	g.Expect(assessment.Complete()).To(gomega.BeFalse())
	g.Expect(assessment.Status()).To(gomega.Equal(StatusStarted))
	assessment.Sections[0].Questions[0].Answers[0].Selected = true
	g.Expect(assessment.Complete()).To(gomega.BeTrue())
	g.Expect(assessment.Status()).To(gomega.Equal(StatusComplete))
}
