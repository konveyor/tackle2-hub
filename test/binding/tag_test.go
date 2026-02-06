package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTag(t *testing.T) {
	g := NewGomegaWithT(t)

	tagCategory := &api.TagCategory{
		Name:  "Test Category for Tag",
		Color: "#00dd00",
	}

	// Get seeded.
	seeded, err := client.Tag.List()
	g.Expect(err).To(BeNil())

	// CREATE: tag category.
	err = client.TagCategory.Create(tagCategory)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.TagCategory.Delete(tagCategory.ID)
	})

	// Define the tag to create
	tag := &api.Tag{
		Name: "Test Linux",
		Category: api.Ref{
			ID:   tagCategory.ID,
			Name: tagCategory.Name,
		},
	}

	// CREATE: Create the tag
	err = client.Tag.Create(tag)
	g.Expect(err).To(BeNil())
	g.Expect(tag.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Tag.Delete(tag.ID)
	})

	// GET: List tags
	list, err := client.Tag.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(tag, list[len(seeded)])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the tag and verify it matches
	retrieved, err := client.Tag.Get(tag.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(tag, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the tag
	tag.Name = "Updated Test Linux"

	err = client.Tag.Update(tag)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Tag.Get(tag.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(tag, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the tag
	err = client.Tag.Delete(tag.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Tag.Get(tag.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestTagEnsure tests idempotent create-or-get operation
func TestTagEnsure(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create tag category for the tag
	tagCategory := &api.TagCategory{
		Name:  "Test Category for Ensure",
		Color: "#00ff00",
	}
	err := client.TagCategory.Create(tagCategory)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.TagCategory.Delete(tagCategory.ID)
	})

	// Define the tag
	tag := &api.Tag{
		Name: "Test Ensure Tag",
		Category: api.Ref{
			ID:   tagCategory.ID,
			Name: tagCategory.Name,
		},
	}

	// ENSURE: First call should create the tag
	err = client.Tag.Ensure(tag)
	g.Expect(err).To(BeNil())
	g.Expect(tag.ID).NotTo(BeZero())
	firstID := tag.ID
	t.Cleanup(func() {
		_ = client.Tag.Delete(tag.ID)
	})

	// GET: Verify tag was created
	retrieved, err := client.Tag.Get(tag.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Name).To(Equal(tag.Name))

	// ENSURE: Second call with same name should return existing tag
	tag2 := &api.Tag{
		Name: "Test Ensure Tag",
		Category: api.Ref{
			ID:   tagCategory.ID,
			Name: tagCategory.Name,
		},
	}
	err = client.Tag.Ensure(tag2)
	g.Expect(err).To(BeNil())
	g.Expect(tag2.ID).To(Equal(firstID), "Ensure should return existing tag with same name")

	// DELETE and ENSURE: Delete then ensure should recreate
	err = client.Tag.Delete(tag.ID)
	g.Expect(err).To(BeNil())

	// Reset ID to 0 and ensure again
	tag.ID = 0
	err = client.Tag.Ensure(tag)
	g.Expect(err).To(BeNil())
	g.Expect(tag.ID).NotTo(BeZero())
	g.Expect(tag.ID).NotTo(Equal(firstID), "Ensured tag after delete should have new ID")

	// Clean up recreated tag
	err = client.Tag.Delete(tag.ID)
	g.Expect(err).To(BeNil())
}
