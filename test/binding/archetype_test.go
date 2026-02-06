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
