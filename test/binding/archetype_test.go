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

	defer func() {
		_ = client.Archetype.Delete(archetype.ID)
	}()

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
