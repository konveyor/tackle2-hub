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

	// Create stakeholder and stakeholder group
	stakeholder := &api.Stakeholder{
		Name:  "Archetype Owner",
		Email: "owner@archetype.local",
	}
	err := client.Stakeholder.Create(stakeholder)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Stakeholder.Delete(stakeholder.ID)
	})

	stakeholderGroup := &api.StakeholderGroup{
		Name:        "Archetype Group",
		Description: "Group for archetype",
	}
	err = client.StakeholderGroup.Create(stakeholderGroup)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.StakeholderGroup.Delete(stakeholderGroup.ID)
	})

	// Define the archetype to create
	archetype := &api.Archetype{
		Name:        "Minimal",
		Description: "Archetype minimal sample 1",
		Comments:    "Archetype comments",
		Tags: []api.TagRef{
			{ID: 1}, // Use seeded tag
			{ID: 2}, // Use seeded tag
		},
		Criteria: []api.TagRef{
			{ID: 3}, // Use seeded tag
			{ID: 4}, // Use seeded tag
		},
		Stakeholders: []api.Ref{
			{ID: stakeholder.ID},
		},
		StakeholderGroups: []api.Ref{
			{ID: stakeholderGroup.ID},
		},
	}

	// CREATE: Create the archetype
	err = client.Archetype.Create(archetype)
	g.Expect(err).To(BeNil())
	g.Expect(archetype.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype.ID)
	})

	// GET: List archetypes
	list, err := client.Archetype.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(archetype, list[0], "Tags.Name", "Criteria.Name", "Stakeholders.Name", "StakeholderGroups.Name")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the archetype and verify it matches
	retrieved, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(archetype, retrieved, "Tags.Name", "Criteria.Name", "Stakeholders.Name", "StakeholderGroups.Name")
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
		"Profiles.CreateTime",
		"Tags.Name",
		"Criteria.Name",
		"Stakeholders.Name",
		"StakeholderGroups.Name")
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

// TestArchetypeApplications tests that applications matching archetype criteria
// are correctly associated and listed in the Applications field.
func TestArchetypeApplications(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create archetype with criteria
	archetype := &api.Archetype{
		Name:        "Test Archetype with Applications",
		Description: "Archetype for testing application associations",
		Criteria: []api.TagRef{
			{ID: 3}, // Use seeded tag as criteria
		},
	}

	// CREATE: Create the archetype
	err := client.Archetype.Create(archetype)
	g.Expect(err).To(BeNil())
	g.Expect(archetype.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype.ID)
	})

	// Create first application matching archetype criteria
	app1 := &api.Application{
		Name:        "Test Application 1",
		Description: "First application matching archetype",
	}
	err = client.Application.Create(app1)
	g.Expect(err).To(BeNil())
	g.Expect(app1.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app1.ID)
	})

	// Add tag to app1 using the tag API
	selected1 := client.Application.Select(app1.ID)
	err = selected1.Tag.Add(3) // Add seeded tag that matches archetype criteria
	g.Expect(err).To(BeNil())

	// Create second application matching archetype criteria
	app2 := &api.Application{
		Name:        "Test Application 2",
		Description: "Second application matching archetype",
	}
	err = client.Application.Create(app2)
	g.Expect(err).To(BeNil())
	g.Expect(app2.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app2.ID)
	})

	// Add tag to app2 using the tag API
	selected2 := client.Application.Select(app2.ID)
	err = selected2.Tag.Add(3) // Add seeded tag that matches archetype criteria
	g.Expect(err).To(BeNil())

	// Create third application matching archetype criteria
	app3 := &api.Application{
		Name:        "Test Application 3",
		Description: "Third application matching archetype",
	}
	err = client.Application.Create(app3)
	g.Expect(err).To(BeNil())
	g.Expect(app3.ID).NotTo(BeZero())
	t.Cleanup(func() {
		_ = client.Application.Delete(app3.ID)
	})

	// Add tag to app3 using the tag API
	selected3 := client.Application.Select(app3.ID)
	err = selected3.Tag.Add(3) // Add seeded tag that matches archetype criteria
	g.Expect(err).To(BeNil())

	// GET: Retrieve archetype and verify Applications field
	retrieved, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())

	// Verify exactly 3 applications are listed
	g.Expect(len(retrieved.Applications)).To(Equal(3), "Archetype should have exactly 3 applications")

	// Build expected application IDs and names
	expectedApps := map[uint]string{
		app1.ID: app1.Name,
		app2.ID: app2.Name,
		app3.ID: app3.Name,
	}

	// Verify each application ref has correct ID and name
	actualApps := make(map[uint]string)
	for _, appRef := range retrieved.Applications {
		actualApps[appRef.ID] = appRef.Name
	}

	// Verify all expected apps are present with correct names
	for expectedID, expectedName := range expectedApps {
		actualName, found := actualApps[expectedID]
		g.Expect(found).To(BeTrue(),
			"Application ID %d (%s) should be in archetype's Applications list",
			expectedID,
			expectedName)
		g.Expect(actualName).To(Equal(expectedName),
			"Application ID %d should have name %s, got %s",
			expectedID,
			expectedName,
			actualName)
	}

	// Verify no duplicate IDs - this specifically tests for the pointer reuse bug
	appIDs := make(map[uint]int)
	for _, appRef := range retrieved.Applications {
		appIDs[appRef.ID]++
	}
	g.Expect(len(appIDs)).To(Equal(3), "Should have exactly 3 unique application IDs")
	for appID, count := range appIDs {
		g.Expect(count).To(Equal(1), "Application ID %d should appear exactly once, not %d times", appID, count)
	}
}

// TestArchetypeAssessedRiskConfidence tests that archetypes correctly compute
// Assessed, Risk, and Confidence fields from their assessments.
func TestArchetypeAssessedRiskConfidence(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a required questionnaire with questions that have risk levels
	questionnaire := &api.Questionnaire{
		Name:        "Test Questionnaire for Archetype Risk",
		Description: "Required questionnaire for testing archetype assessment fields",
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
				Name:  "Risk Assessment",
				Questions: []api.Question{
					{
						Order:       1,
						Text:        "What is the application's complexity?",
						Explanation: "Assess technical complexity",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Low complexity",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "Medium complexity",
								Risk:  "yellow",
							},
							{
								Order: 3,
								Text:  "High complexity",
								Risk:  "red",
							},
						},
					},
					{
						Order:       2,
						Text:        "What is the migration effort?",
						Explanation: "Estimate migration work",
						Answers: []api.Answer{
							{
								Order: 1,
								Text:  "Minimal effort",
								Risk:  "green",
							},
							{
								Order: 2,
								Text:  "Significant effort",
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

	// Create archetype for testing
	archetype := &api.Archetype{
		Name:        "Test Archetype for Assessment Fields",
		Description: "Archetype for testing Assessed, Risk, Confidence",
	}
	err = client.Archetype.Create(archetype)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype.ID)
	})

	// GET: Initially archetype should not be assessed
	retrieved, err := client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Assessed).To(BeFalse(), "Archetype should not be assessed without assessments")

	// Create assessment for the archetype
	selected := client.Archetype.Select(archetype.ID)
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
	t.Cleanup(func() {
		_ = client.Assessment.Delete(assessment.ID)
	})

	// GET: With incomplete assessment, archetype should still not be assessed
	retrieved, err = client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Assessed).To(BeFalse(), "Archetype should not be assessed with incomplete assessment")

	// Answer questions - select "Medium complexity" (yellow) and "Minimal effort" (green)
	assessment.Sections[0].Questions[0].Answers[1].Selected = true // Medium complexity (yellow)
	assessment.Sections[0].Questions[1].Answers[0].Selected = true // Minimal effort (green)
	assessment.Status = "complete"

	err = client.Assessment.Update(assessment)
	g.Expect(err).To(BeNil())

	// GET: With completed required assessment, archetype should be assessed
	retrieved, err = client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Assessed).To(BeTrue(), "Archetype should be assessed with completed required assessment")

	// Verify Risk is computed correctly (yellow + green = yellow overall)
	g.Expect(retrieved.Risk).To(Equal("yellow"), "Archetype risk should be yellow based on assessment answers")

	// Verify Confidence is computed (should be non-zero)
	g.Expect(retrieved.Confidence).To(BeNumerically(">", 0), "Archetype confidence should be computed from assessment")

	// Update assessment to all green answers
	assessment.Sections[0].Questions[0].Answers[1].Selected = false // Deselect Medium complexity
	assessment.Sections[0].Questions[0].Answers[0].Selected = true  // Select Low complexity (green)

	err = client.Assessment.Update(assessment)
	g.Expect(err).To(BeNil())

	// GET: Verify risk changed to green
	retrieved, err = client.Archetype.Get(archetype.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Assessed).To(BeTrue())
	g.Expect(retrieved.Risk).To(Equal("green"), "Archetype risk should be green with all green answers")
}

// TestArchetypeCriteriaMatching tests edge cases in archetype-application matching
// based on criteria tags.
func TestArchetypeCriteriaMatching(t *testing.T) {
	g := NewGomegaWithT(t)

	// Test 1: Multiple criteria tags - app must have ALL to match
	archetype1 := &api.Archetype{
		Name:        "Multi-Criteria Archetype",
		Description: "Requires multiple tags to match",
		Criteria: []api.TagRef{
			{ID: 1}, // Use seeded tags
			{ID: 2},
		},
	}
	err := client.Archetype.Create(archetype1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype1.ID)
	})

	// Create app with only ONE of the required tags (partial match)
	appPartial := &api.Application{
		Name:        "Partial Match App",
		Description: "Has only one of two required tags",
	}
	err = client.Application.Create(appPartial)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(appPartial.ID)
	})

	selectedPartial := client.Application.Select(appPartial.ID)
	err = selectedPartial.Tag.Add(1) // Add only tag 1, not tag 2
	g.Expect(err).To(BeNil())

	// Create app with BOTH required tags (full match)
	appFull := &api.Application{
		Name:        "Full Match App",
		Description: "Has both required tags",
	}
	err = client.Application.Create(appFull)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(appFull.ID)
	})

	selectedFull := client.Application.Select(appFull.ID)
	err = selectedFull.Tag.Add(1) // Add tag 1
	g.Expect(err).To(BeNil())
	err = selectedFull.Tag.Add(2) // Add tag 2
	g.Expect(err).To(BeNil())

	// GET: Verify only the app with ALL criteria appears
	retrieved1, err := client.Archetype.Get(archetype1.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(retrieved1.Applications)).To(Equal(1), "Only app with all criteria tags should match")
	g.Expect(retrieved1.Applications[0].ID).To(Equal(appFull.ID), "Full match app should be in Applications")

	// Test 2: Empty criteria - all apps with tags should match
	archetype2 := &api.Archetype{
		Name:        "No Criteria Archetype",
		Description: "Has no criteria, should match all apps with tags",
		Criteria:    []api.TagRef{}, // Empty criteria
	}
	err = client.Archetype.Create(archetype2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype2.ID)
	})

	// Create app with an unrelated tag (not in any criteria)
	appUnrelatedTag := &api.Application{
		Name:        "Unrelated Tag App",
		Description: "Application with tag not in any criteria",
	}
	err = client.Application.Create(appUnrelatedTag)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(appUnrelatedTag.ID)
	})

	selectedUnrelated := client.Application.Select(appUnrelatedTag.ID)
	err = selectedUnrelated.Tag.Add(4) // Add tag 4 (not used in archetype2 criteria)
	g.Expect(err).To(BeNil())

	// GET: Archetype with no criteria should match apps with any tags
	retrieved2, err := client.Archetype.Get(archetype2.ID)
	g.Expect(err).To(BeNil())
	g.Expect(len(retrieved2.Applications)).To(BeNumerically(">=", 1),
		"Archetype with no criteria should match apps with tags")

	// Verify the app with unrelated tag is included
	foundUnrelated := false
	for _, app := range retrieved2.Applications {
		if app.ID == appUnrelatedTag.ID {
			foundUnrelated = true
			break
		}
	}
	g.Expect(foundUnrelated).To(BeTrue(), "App with any tag should match archetype with no criteria")

	// Test 3: App loses matching tag - should disappear from Applications
	archetype3 := &api.Archetype{
		Name:        "Single Criteria Archetype",
		Description: "Requires single tag",
		Criteria: []api.TagRef{
			{ID: 3}, // Use seeded tag
		},
	}
	err = client.Archetype.Create(archetype3)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Archetype.Delete(archetype3.ID)
	})

	// Create app with matching tag
	appDynamic := &api.Application{
		Name:        "Dynamic Match App",
		Description: "Will gain and lose matching tag",
	}
	err = client.Application.Create(appDynamic)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(appDynamic.ID)
	})

	selectedDynamic := client.Application.Select(appDynamic.ID)
	err = selectedDynamic.Tag.Add(3) // Add matching tag
	g.Expect(err).To(BeNil())

	// GET: App should appear in Applications
	retrieved3, err := client.Archetype.Get(archetype3.ID)
	g.Expect(err).To(BeNil())
	foundBefore := false
	for _, app := range retrieved3.Applications {
		if app.ID == appDynamic.ID {
			foundBefore = true
			break
		}
	}
	g.Expect(foundBefore).To(BeTrue(), "App with matching tag should appear in Applications")

	// Remove the matching tag
	err = selectedDynamic.Tag.Delete(3)
	g.Expect(err).To(BeNil())

	// GET: App should disappear from Applications
	retrieved3After, err := client.Archetype.Get(archetype3.ID)
	g.Expect(err).To(BeNil())
	foundAfter := false
	for _, app := range retrieved3After.Applications {
		if app.ID == appDynamic.ID {
			foundAfter = true
			break
		}
	}
	g.Expect(foundAfter).To(BeFalse(), "App without matching tag should disappear from Applications")
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
