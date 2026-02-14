package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTracker(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an identity for the tracker to reference
	identity := &api.Identity{
		Name: "tracker-identity",
		Kind: "basic-auth",
		User: "tracker-user",
		Password: "tracker-password",
	}
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// Define the tracker to create
	tracker := &api.Tracker{
		Name:     "Test Tracker",
		URL:      "https://konveyor.io/test/api/tracker",
		Kind:     "jira-onprem",
		Message:  "Test tracker description",
		Insecure: false,
		Identity: api.Ref{
			ID:   identity.ID,
			Name: identity.Name,
		},
	}

	// CREATE: Create the tracker
	err = client.Tracker.Create(tracker)
	g.Expect(err).To(BeNil())
	g.Expect(tracker.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Tracker.Delete(tracker.ID)
	})

	// GET: List trackers
	list, err := client.Tracker.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(tracker, list[0], "LastUpdated")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the tracker and verify it matches
	retrieved, err := client.Tracker.Get(tracker.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(tracker, retrieved, "LastUpdated")
	g.Expect(eq).To(BeTrue(), report)

	// UPDATE: Modify the tracker
	tracker.Name = "Updated Test Tracker"
	tracker.URL = "https://konveyor.io/test/api/tracker-updated"
	tracker.Message = "Updated tracker description"

	err = client.Tracker.Update(tracker)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Tracker.Get(tracker.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report = cmp.Eq(tracker, updated, "UpdateUser", "LastUpdated")
	g.Expect(eq).To(BeTrue(), report)

	// DELETE: Remove the tracker
	err = client.Tracker.Delete(tracker.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Tracker.Get(tracker.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
