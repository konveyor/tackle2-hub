package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTagCategory(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the tag category to create
	tagCategory := &api.TagCategory{
		Name:  "Test OS",
		Color: "#dd0000",
	}

	// Get seeded.
	seeded, err := client.TagCategory.List()
	g.Expect(err).To(BeNil())

	// CREATE: Create the tag category
	err = client.TagCategory.Create(tagCategory)
	g.Expect(err).To(BeNil())
	g.Expect(tagCategory.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.TagCategory.Delete(tagCategory.ID)
	})

	// GET: List tag categories
	list, err := client.TagCategory.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded) + 1))
	eq, report := cmp.Eq(tagCategory, list[len(seeded)])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the tag category and verify it matches
	retrieved, err := client.TagCategory.Get(tagCategory.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(tagCategory, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the tag category
	tagCategory.Name = "Updated Test OS"
	tagCategory.Color = "#ee0000"

	err = client.TagCategory.Update(tagCategory)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.TagCategory.Get(tagCategory.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(tagCategory, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the tag category
	err = client.TagCategory.Delete(tagCategory.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.TagCategory.Get(tagCategory.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}

// TestTagCategoryEnsure tests idempotent create-or-get operation
func TestTagCategoryEnsure(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the tag category
	tagCategory := &api.TagCategory{
		Name:  "Test Ensure Category",
		Color: "#ff00ff",
	}

	// ENSURE: First call should create the tag category
	err := client.TagCategory.Ensure(tagCategory)
	g.Expect(err).To(BeNil())
	g.Expect(tagCategory.ID).NotTo(BeZero())
	firstID := tagCategory.ID
	t.Cleanup(func() {
		_ = client.TagCategory.Delete(tagCategory.ID)
	})

	// GET: Verify tag category was created
	retrieved, err := client.TagCategory.Get(tagCategory.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved.Name).To(Equal(tagCategory.Name))

	// ENSURE: Second call with same name should return existing tag category
	tagCategory2 := &api.TagCategory{
		Name:  "Test Ensure Category",
		Color: "#ff00ff",
	}
	err = client.TagCategory.Ensure(tagCategory2)
	g.Expect(err).To(BeNil())
	g.Expect(tagCategory2.ID).To(Equal(firstID), "Ensure should return existing tag category with same name")

	// DELETE and ENSURE: Delete then ensure should recreate
	err = client.TagCategory.Delete(tagCategory.ID)
	g.Expect(err).To(BeNil())

	// Reset ID to 0 and ensure again
	tagCategory.ID = 0
	err = client.TagCategory.Ensure(tagCategory)
	g.Expect(err).To(BeNil())
	g.Expect(tagCategory.ID).NotTo(BeZero())
	g.Expect(tagCategory.ID).NotTo(Equal(firstID), "Ensured tag category after delete should have new ID")

	// Clean up recreated tag category
	err = client.TagCategory.Delete(tagCategory.ID)
	g.Expect(err).To(BeNil())
}
