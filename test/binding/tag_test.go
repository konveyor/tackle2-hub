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

	// Create a tag category for the tag to reference
	tagCategory := &api.TagCategory{
		Name:  "Test Category for Tag",
		Color: "#00dd00",
	}
	err := client.TagCategory.Create(tagCategory)
	g.Expect(err).To(BeNil())
	defer func() {
		_ = client.TagCategory.Delete(tagCategory.ID)
	}()

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

	defer func() {
		_ = client.Tag.Delete(tag.ID)
	}()

	// GET: Retrieve the tag and verify it matches
	retrieved, err := client.Tag.Get(tag.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(tag, retrieved)
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
