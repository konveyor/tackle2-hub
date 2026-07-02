package binding

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestPermission(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get seeded
	seeded, err := client.Scope.List()
	g.Expect(err).To(BeNil())

	// GET: List scopes
	list, err := client.Scope.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(len(seeded)))
}
