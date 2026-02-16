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

// TestDependencyReverse tests circular dependency prevention
func TestDependencyReverse(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create three applications for dependency chain testing
	app1 := &api.Application{
		Name:        "First",
		Description: "First application",
	}
	err := client.Application.Create(app1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app1.ID)
	})

	app2 := &api.Application{
		Name:        "Second",
		Description: "Second application",
	}
	err = client.Application.Create(app2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app2.ID)
	})

	app3 := &api.Application{
		Name:        "Third",
		Description: "Third application",
	}
	err = client.Application.Create(app3)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(app3.ID)
	})

	// CREATE: First dependency App1 -> App2 (should succeed)
	dep1 := &api.Dependency{
		From: api.Ref{ID: app1.ID},
		To:   api.Ref{ID: app2.ID},
	}
	err = client.Dependency.Create(dep1)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Dependency.Delete(dep1.ID)
	})

	// CREATE: Second dependency App2 -> App3 (should succeed)
	dep2 := &api.Dependency{
		From: api.Ref{ID: app2.ID},
		To:   api.Ref{ID: app3.ID},
	}
	err = client.Dependency.Create(dep2)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Dependency.Delete(dep2.ID)
	})

	// CREATE: Indirect reverse dependency App3 -> App1 (should FAIL)
	// This would create a circular dependency: App1 -> App2 -> App3 -> App1
	indirectReverse := &api.Dependency{
		From: api.Ref{ID: app3.ID},
		To:   api.Ref{ID: app1.ID},
	}
	err = client.Dependency.Create(indirectReverse)
	g.Expect(err).NotTo(BeNil(), "Indirect reverse dependency should be prevented")

	// CREATE: Third dependency App1 -> App3 (should succeed)
	// This is allowed because it doesn't create a cycle
	dep3 := &api.Dependency{
		From: api.Ref{ID: app1.ID},
		To:   api.Ref{ID: app3.ID},
	}
	err = client.Dependency.Create(dep3)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Dependency.Delete(dep3.ID)
	})

	// CREATE: Direct reverse dependency App2 -> App1 (should FAIL)
	// This would reverse the existing App1 -> App2 dependency
	directReverse1 := &api.Dependency{
		From: api.Ref{ID: app2.ID},
		To:   api.Ref{ID: app1.ID},
	}
	err = client.Dependency.Create(directReverse1)
	g.Expect(err).NotTo(BeNil(), "Direct reverse dependency App2 -> App1 should be prevented")

	// CREATE: Another direct reverse dependency App3 -> App2 (should FAIL)
	// This would reverse the existing App2 -> App3 dependency
	directReverse2 := &api.Dependency{
		From: api.Ref{ID: app3.ID},
		To:   api.Ref{ID: app2.ID},
	}
	err = client.Dependency.Create(directReverse2)
	g.Expect(err).NotTo(BeNil(), "Direct reverse dependency App3 -> App2 should be prevented")
}
