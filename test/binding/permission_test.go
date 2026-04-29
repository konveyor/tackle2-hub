package binding

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestPermission(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get seeded
	seeded, err := client.Permission.List()
	g.Expect(err).To(BeNil())

	// GET: List permissions
	list, err := client.Permission.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded)))

	// GET: Retrieve the permission and verify it matches
	retrieved, err := client.Permission.Get(seeded[0].ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report := cmp.Eq(seeded[0], retrieved)
	g.Expect(eq).To(BeTrue(), report)
}
