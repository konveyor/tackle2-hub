package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestDependency(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create two applications for the dependency to reference
	appFrom := &api.Application{
		Name:        "Gateway App",
		Description: "Gateway application",
	}
	err := client.Application.Create(appFrom)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(appFrom.ID)
	})

	appTo := &api.Application{
		Name:        "Inventory App",
		Description: "Inventory application",
	}
	err = client.Application.Create(appTo)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(appTo.ID)
	})

	// Define the dependency to create
	dependency := &api.Dependency{
		From: api.Ref{
			ID:   appFrom.ID,
			Name: appFrom.Name,
		},
		To: api.Ref{
			ID:   appTo.ID,
			Name: appTo.Name,
		},
	}

	// CREATE: Create the dependency
	err = client.Dependency.Create(dependency)
	g.Expect(err).To(BeNil())
	g.Expect(dependency.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Dependency.Delete(dependency.ID)
	})

	// GET: List dependencies
	list, err := client.Dependency.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(dependency, list[0])
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the dependency and verify it matches
	retrieved, err := client.Dependency.Get(dependency.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(dependency, retrieved)
	g.Expect(eq).To(BeTrue(), report)

	// NOTE: Dependency does not have an Update method according to the API

	// DELETE: Remove the dependency
	err = client.Dependency.Delete(dependency.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Dependency.Get(dependency.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
