package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestTicket(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create an identity for the tracker
	identity := &api.Identity{
		Name:     "ticket-tracker-identity",
		Kind:     "basic-auth",
		User:     "ticket-user",
		Password: "ticket-password",
	}
	err := client.Identity.Create(identity)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	// Create a tracker for the ticket to reference
	tracker := &api.Tracker{
		Name:     "Test Ticket Tracker",
		URL:      "https://konveyor.io/test/api/ticket-tracker",
		Kind:     "jira-onprem",
		Message:  "Test ticket tracker",
		Insecure: false,
		Identity: api.Ref{
			ID:   identity.ID,
			Name: identity.Name,
		},
	}
	err = client.Tracker.Create(tracker)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Tracker.Delete(tracker.ID)
	})

	// Create an application for the ticket to reference
	application := &api.Application{
		Name:        "Test Ticket App",
		Description: "Application for ticket testing",
	}
	err = client.Application.Create(application)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Application.Delete(application.ID)
	})

	// Define the ticket to create
	ticket := &api.Ticket{
		Kind:   "task",
		Parent: "PROJECT-1",
		Application: api.Ref{
			ID:   application.ID,
			Name: application.Name,
		},
		Tracker: api.Ref{
			ID:   tracker.ID,
			Name: tracker.Name,
		},
	}

	// CREATE: Create the ticket
	err = client.Ticket.Create(ticket)
	g.Expect(err).To(BeNil())
	g.Expect(ticket.ID).NotTo(BeZero())

	t.Cleanup(func() {
		_ = client.Ticket.Delete(ticket.ID)
	})

	// GET: List tickets
	list, err := client.Ticket.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(Equal(1))
	eq, report := cmp.Eq(ticket, list[0], "Status", "Message", "LastUpdated", "Fields", "Link")
	g.Expect(eq).To(BeTrue(), report)

	// GET: Retrieve the ticket and verify it matches
	retrieved, err := client.Ticket.Get(ticket.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())
	eq, report = cmp.Eq(ticket, retrieved, "Status", "Message", "LastUpdated", "Fields", "Link")
	g.Expect(eq).To(BeTrue(), report)

	// NOTE: Ticket does not have an Update method according to the API

	// DELETE: Remove the ticket
	err = client.Ticket.Delete(ticket.ID)
	g.Expect(err).To(BeNil())

	// Verify deletion - Get should fail
	_, err = client.Ticket.Get(ticket.ID)
	g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
}
