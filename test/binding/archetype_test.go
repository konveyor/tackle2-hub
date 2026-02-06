package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestArchetype(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the archetype to create
	archetype := &api.Archetype{
		Name:        "Minimal",
		Description: "Archetype minimal sample 1",
		Comments:    "Archetype comments",
	}

	// CREATE: Create the archetype
	err := client.Archetype.Create(archetype)
	g.Expect(err).To(BeNil())
	g.Expect(archetype.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype.ID)
	})

	// GET: List archetypes
	list, err := client.Archetype.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(archetype, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the archetype and verify it matches
	retrieved, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(archetype, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the archetype
	archetype.Name = "Updated Minimal"
	archetype.Description = "Updated archetype description"
	archetype.Comments = "Updated comments"
	archetype.Profiles = []api.TargetProfile{
		{Name: "openshift"},
	}

	err = client.Archetype.Update(archetype)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(
		archetype,
		updated,
		"UpdateUser",
		"Profiles.ID",
		"Profiles.CreateTime")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the archetype
	err = client.Archetype.Delete(archetype.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Archetype.Get(archetype.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestArchetypeProfileManagement tests complex profile and generator management
func TestArchetypeProfileManagement(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create generators
	genA := &api.Generator{Name: "genA", Kind: "helm"}
	err := client.Generator.Create(genA)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Generator.Delete(genA.ID)
	})

	genB := &api.Generator{Name: "genB", Kind: "helm"}
	err = client.Generator.Create(genB)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Generator.Delete(genB.ID)
	})

	genC := &api.Generator{Name: "genC", Kind: "helm"}
	err = client.Generator.Create(genC)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Generator.Delete(genC.ID)
	})

	genD := &api.Generator{Name: "genD", Kind: "helm"}
	err = client.Generator.Create(genD)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Generator.Delete(genD.ID)
	})

	// Create analysis profile
	analysisProfile := &api.AnalysisProfile{
		Name:        "Test Analysis Profile for Archetype",
		Description: "Profile for testing archetype profiles",
	}
	err = client.AnalysisProfile.Create(analysisProfile)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.AnalysisProfile.Delete(analysisProfile.ID)
	})

	// CREATE: Create archetype with initial profile and generator
	archetype := &api.Archetype{
		Name:        "Test Archetype with Profiles",
		Description: "Archetype for testing profile management",
		Profiles: []api.TargetProfile{
			{
				Name: "initial-profile",
				AnalysisProfile: &api.Ref{
					ID:   analysisProfile.ID,
					Name: analysisProfile.Name,
				},
				Generators: []api.Ref{
					{ID: genA.ID, Name: genA.Name},
				},
			},
		},
	}
	err = client.Archetype.Create(archetype)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype.ID)
	})

	// GET: Verify creation
	retrieved, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(retrieved.Profiles)).To(Equal(1))
	g.Expect(retrieved.Profiles[0].Name).To(Equal("initial-profile"))
	g.Expect(len(retrieved.Profiles[0].Generators)).To(Equal(1))
	initialProfileCount := len(retrieved.Profiles)

	// UPDATE: Add a new profile with a different generator
	archetype.Name += "-Updated"
	archetype.Profiles = append(
		archetype.Profiles,
		api.TargetProfile{
			Name: "Added",
			Generators: []api.Ref{
				{ID: genD.ID, Name: genD.Name},
			},
		})
	err = client.Archetype.Update(archetype)
	g.Expect(err).To(BeNil())

	// GET: Verify profile was added
	updated, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated.Name).To(Equal(archetype.Name))
	g.Expect(len(updated.Profiles)).To(Equal(initialProfileCount + 1))

	// UPDATE: Add more generators to all profiles and verify ordering
	for i := range archetype.Profiles {
		p := &archetype.Profiles[i]
		p.Generators = append(
			p.Generators,
			api.Ref{ID: genC.ID, Name: genC.Name},
			api.Ref{ID: genB.ID, Name: genB.Name},
		)
	}
	err = client.Archetype.Update(archetype)
	g.Expect(err).To(BeNil())

	// GET: Verify generator ordering is preserved
	finalArchetype, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(finalArchetype.Profiles)).To(Equal(2))

	// Verify first profile has generators in order: genA, genC, genB
	profile0 := finalArchetype.Profiles[0]
	g.Expect(len(profile0.Generators)).To(Equal(3))
	g.Expect(profile0.Generators[0].ID).To(Equal(genA.ID))
	g.Expect(profile0.Generators[1].ID).To(Equal(genC.ID))
	g.Expect(profile0.Generators[2].ID).To(Equal(genB.ID))

	// Verify second profile has generators in order: genD, genC, genB
	profile1 := finalArchetype.Profiles[1]
	g.Expect(len(profile1.Generators)).To(Equal(3))
	g.Expect(profile1.Generators[0].ID).To(Equal(genD.ID))
	g.Expect(profile1.Generators[1].ID).To(Equal(genC.ID))
	g.Expect(profile1.Generators[2].ID).To(Equal(genB.ID))
}

// TestArchetypeAssessment tests the Archetype.Select().Assessment subresource
func TestArchetypeAssessment(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a questionnaire for the assessment
	questionnaire := &api.Questionnaire{
		Name:        "Test Questionnaire for Archetype",
		Description: "Questionnaire for testing archetype assessments",
		Required:    true,
		Thresholds: api.Thresholds{
			Red:     30,
			Yellow:  20,
			Unknown: 10,
		},
		RiskMessages: api.RiskMessages{
			Red:     "High risk",
			Yellow:  "Medium risk",
			Green:   "Low risk",
			Unknown: "Unknown risk",
		},
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Test Section",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "Test question?",
						Explanation: "Test explanation",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Answer 1",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "Answer 2",
								Risk:  "red",
							},
						},
					},
				},
			},
		},
	}
	err := client.Questionnaire.Create(questionnaire)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Questionnaire.Delete(questionnaire.ID)
	})

	// Create an archetype for testing
	archetype := &api.Archetype{
		Name:        "Test Archetype for Assessment",
		Description: "Archetype for testing assessment subresource",
	}
	err = client.Archetype.Create(archetype)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype.ID)
	})

	// Get the selected archetype API
	selected := client.Archetype.Select(archetype.ID)

	// CREATE: Create an assessment for the archetype
	assessment := &api.Assessment{
		Questionnaire: api.Ref{
			ID:   questionnaire.ID,
			Name: questionnaire.Name,
		},
		Archetype: &api.Ref{
			ID:   archetype.ID,
			Name: archetype.Name,
		},
		Status: "started",
	}
	err = selected.Assessment.Create(assessment)
	g.Expect(err).To(BeNil())
	g.Expect(assessment.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment.ID)
	})

	// LIST: Verify assessment was created for the archetype
	list, err := selected.Assessment.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	g.Expect(list[0].ID).To(Equal(assessment.ID))
	g.Expect(list[0].Archetype).NotTo(BeNil())
	g.Expect(list[0].Archetype.ID).To(Equal(archetype.ID))

	// Create a second questionnaire for the second assessment
	// Note: UNIQUE constraint on (ArchetypeID, QuestionnaireID) requires different questionnaire
	questionnaire2 := &api.Questionnaire{
		Name:        "Test Questionnaire 2 for Archetype",
		Description: "Second questionnaire for testing archetype assessments",
		Required:    false,
		Thresholds: api.Thresholds{
			Red:     25,
			Yellow:  15,
			Unknown: 5,
		},
		RiskMessages: api.RiskMessages{
			Red:     "High risk",
			Yellow:  "Medium risk",
			Green:   "Low risk",
			Unknown: "Unknown risk",
		},
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Test Section 2",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "Another test question?",
						Explanation: "Another test explanation",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Answer A",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "Answer B",
								Risk:  "yellow",
							},
						},
					},
				},
			},
		},
	}
	err = client.Questionnaire.Create(questionnaire2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Questionnaire.Delete(questionnaire2.ID)
	})

	// CREATE: Create a second assessment with different questionnaire
	assessment2 := &api.Assessment{
		Questionnaire: api.Ref{
			ID:   questionnaire2.ID,
			Name: questionnaire2.Name,
		},
		Archetype: &api.Ref{
			ID:   archetype.ID,
			Name: archetype.Name,
		},
		Status: "complete",
	}
	err = selected.Assessment.Create(assessment2)
	g.Expect(err).To(BeNil())
	g.Expect(assessment2.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment2.ID)
	})

	// LIST: Verify both assessments are associated with the archetype
	list, err = selected.Assessment.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(2))

	// Verify all assessments belong to this archetype
	for _, a := range list {
		g.Expect(a.Archetype).NotTo(BeNil())
		g.Expect(a.Archetype.ID).To(Equal(archetype.ID))
	}
}

// TestArchetypeAssessmentMultiple tests assessments across multiple archetypes
func TestArchetypeAssessmentMultiple(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a questionnaire
	questionnaire := &api.Questionnaire{
		Name:        "Test Questionnaire for Multiple Archetypes",
		Description: "Questionnaire for testing multiple archetypes",
		Required:    true,
		Thresholds: api.Thresholds{
			Red:     30,
			Yellow:  20,
			Unknown: 10,
		},
		RiskMessages: api.RiskMessages{
			Red:     "High risk",
			Yellow:  "Medium risk",
			Green:   "Low risk",
			Unknown: "Unknown risk",
		},
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Test Section",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "Test question?",
						Explanation: "Test explanation",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Answer 1",
								Risk:  "green",
							},
						},
					},
				},
			},
		},
	}
	err := client.Questionnaire.Create(questionnaire)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Questionnaire.Delete(questionnaire.ID)
	})

	// Create first archetype
	archetype1 := &api.Archetype{
		Name:        "Test Archetype 1",
		Description: "First archetype for testing",
	}
	err = client.Archetype.Create(archetype1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype1.ID)
	})

	// Create second archetype
	archetype2 := &api.Archetype{
		Name:        "Test Archetype 2",
		Description: "Second archetype for testing",
	}
	err = client.Archetype.Create(archetype2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype2.ID)
	})

	// Create second questionnaire for testing multiple assessments per archetype
	questionnaire2 := &api.Questionnaire{
		Name:        "Test Questionnaire 2 for Multiple Archetypes",
		Description: "Second questionnaire for testing",
		Required:    false,
		Thresholds: api.Thresholds{
			Red:     25,
			Yellow:  15,
			Unknown: 5,
		},
		RiskMessages: api.RiskMessages{
			Red:     "High risk",
			Yellow:  "Medium risk",
			Green:   "Low risk",
			Unknown: "Unknown risk",
		},
		Sections: []api.Section{
			{
				Order: 1,
				Name:  "Test Section 2",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "Another question?",
						Explanation: "Another explanation",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Answer A",
								Risk:  "yellow",
							},
						},
					},
				},
			},
		},
	}
	err = client.Questionnaire.Create(questionnaire2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Questionnaire.Delete(questionnaire2.ID)
	})

	// Create assessments for first archetype
	selected1 := client.Archetype.Select(archetype1.ID)
	assessment1a := &api.Assessment{
		Questionnaire: api.Ref{
			ID:   questionnaire.ID,
			Name: questionnaire.Name,
		},
		Archetype: &api.Ref{
			ID:   archetype1.ID,
			Name: archetype1.Name,
		},
		Status: "started",
	}
	err = selected1.Assessment.Create(assessment1a)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment1a.ID)
	})

	// Second assessment for archetype1 uses different questionnaire
	assessment1b := &api.Assessment{
		Questionnaire: api.Ref{
			ID:   questionnaire2.ID,
			Name: questionnaire2.Name,
		},
		Archetype: &api.Ref{
			ID:   archetype1.ID,
			Name: archetype1.Name,
		},
		Status: "complete",
	}
	err = selected1.Assessment.Create(assessment1b)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment1b.ID)
	})

	// Create assessment for second archetype
	selected2 := client.Archetype.Select(archetype2.ID)
	assessment2a := &api.Assessment{
		Questionnaire: api.Ref{
			ID:   questionnaire.ID,
			Name: questionnaire.Name,
		},
		Archetype: &api.Ref{
			ID:   archetype2.ID,
			Name: archetype2.Name,
		},
		Status: "started",
	}
	err = selected2.Assessment.Create(assessment2a)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment2a.ID)
	})

	// LIST: Verify first archetype has 2 assessments
	list1, err := selected1.Assessment.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list1)).To(Equal(2))
	for _, a := range list1 {
		g.Expect(a.Archetype.ID).To(Equal(archetype1.ID))
	}

	// LIST: Verify second archetype has 1 assessment
	list2, err := selected2.Assessment.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list2)).To(Equal(1))
	g.Expect(list2[0].Archetype.ID).To(Equal(archetype2.ID))
}
