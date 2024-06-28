package assessment

import (
	"testing"

	"github.com/konveyor/tackle2-hub/model"
	"github.com/onsi/gomega"
)

func TestPrepare(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	assessment := model.Assessment{}
	assessment.Sections = []model.Section{
		{
			Questions: []model.Question{
				{
					Text: "Default",
					Answers: []model.Answer{
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
					IncludeFor: []model.CategorizedTag{
						{Category: "Category", Tag: "Tag"},
					},
					Answers: []model.Answer{
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
					ExcludeFor: []model.CategorizedTag{
						{Category: "Category", Tag: "Tag"},
					},
					Answers: []model.Answer{
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
					Answers: []model.Answer{
						{
							Text: "Answer1",
							AutoAnswerFor: []model.CategorizedTag{
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
	a := Assessment{}
	a.With(&assessment)

	tagResolver := TagResolver{
		cache: map[string]map[string]*model.Tag{
			"Category": {"Tag": {Model: model.Model{ID: 1}}},
		},
	}
	tags := NewSet()
	tags.Add(1)

	a.Prepare(&tagResolver, tags)
	questions := a.Sections[0].Questions

	g.Expect(len(questions)).To(gomega.Equal(3))
	g.Expect(questions[0].Text).To(gomega.Equal("Default"))
	g.Expect(a.questionAnswered(&questions[0])).To(gomega.BeFalse())
	g.Expect(questions[1].Text).To(gomega.Equal("Should Include"))
	g.Expect(a.questionAnswered(&questions[1])).To(gomega.BeFalse())
	g.Expect(questions[2].Text).To(gomega.Equal("AutoAnswer"))
	g.Expect(a.questionAnswered(&questions[2])).To(gomega.BeTrue())
	g.Expect(questions[2].Answers[0].Text).To(gomega.Equal("Answer1"))
	g.Expect(questions[2].Answers[0].AutoAnswered).To(gomega.BeTrue())
	g.Expect(questions[2].Answers[0].Selected).To(gomega.BeTrue())
}
