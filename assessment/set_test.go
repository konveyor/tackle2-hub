package assessment

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestSet_Superset(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	a := NewSet()
	b := NewSet()
	a.Add(1, 2, 3, 4)
	b.Add(1, 2, 3, 4)

	g.Expect(a.Superset(b, false)).To(gomega.BeTrue())
	g.Expect(b.Superset(a, false)).To(gomega.BeTrue())
	g.Expect(a.Superset(b, true)).To(gomega.BeFalse())
	g.Expect(b.Superset(a, true)).To(gomega.BeFalse())

	a.Add(5)
	g.Expect(a.Superset(b, false)).To(gomega.BeTrue())
	g.Expect(a.Superset(b, true)).To(gomega.BeTrue())
	g.Expect(b.Superset(a, false)).To(gomega.BeFalse())
	g.Expect(b.Superset(a, true)).To(gomega.BeFalse())
}

func TestSet_Subset(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	a := NewSet()
	b := NewSet()
	a.Add(1, 2, 3, 4)
	b.Add(1, 2, 3, 4)

	g.Expect(a.Subset(b, false)).To(gomega.BeTrue())
	g.Expect(b.Subset(a, false)).To(gomega.BeTrue())
	g.Expect(a.Subset(b, true)).To(gomega.BeFalse())
	g.Expect(b.Subset(a, true)).To(gomega.BeFalse())

	b.Add(5)
	g.Expect(a.Subset(b, false)).To(gomega.BeTrue())
	g.Expect(a.Subset(b, true)).To(gomega.BeTrue())
	g.Expect(b.Subset(a, false)).To(gomega.BeFalse())
	g.Expect(b.Subset(a, true)).To(gomega.BeFalse())
}
