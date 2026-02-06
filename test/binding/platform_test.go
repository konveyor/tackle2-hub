package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestPlatform(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the platform to create
	platform := &api.Platform{
		Kind: "test-platform",
		Name: "Test Platform",
		URL:  "http://localhost:8080",
	}

	// CREATE: Create the platform
	err := client.Platform.Create(platform)
	g.Expect(err).To(BeNil())
	g.Expect(platform.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Platform.Delete(platform.ID)
	})

	// GET: List platforms
	list, err := client.Platform.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(platform, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the platform and verify it matches
	retrieved, err := client.Platform.Get(platform.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(platform, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the platform
	platform.Name = "Updated Test Platform"
	platform.URL = "http://localhost:9090"

	err = client.Platform.Update(platform)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Platform.Get(platform.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(platform, updated, "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the platform
	err = client.Platform.Delete(platform.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Platform.Get(platform.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
