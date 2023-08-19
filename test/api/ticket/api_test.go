package ticket

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTicketCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("Ticket "+r.Kind+" CRUD", func(t *testing.T) {

			// Create a sample Application for the ticket.
			app := api.Application{
				Name: r.Application.Name,
			}
			assert.Must(t, Application.Create(&app))

			// Create a sample identity for the tracker
			identity := api.Identity{
				Name: "Sample Identity",
				Kind: "Sample Kind",
			}
			assert.Must(t, Identity.Create(&identity))

			// create a sample tracker for the ticket.
			tracker := api.Tracker{
				Name: r.Tracker.Name,
				URL:  "www.sample.com",
				Kind: "jira-onprem",
				Identity: api.Ref{
					ID:   1,
					Name: "Sample Identity",
				},
			}
			assert.Must(t, Tracker.Create(&tracker))

			// Create a sample ticket
			assert.Must(t, Ticket.Create(&r))

			// Get.
			got, err := Ticket.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare got values with expected values.
			AssertEqualTickets(t, got, r)

			// Delete identity and tracker.
			assert.Must(t, Ticket.Delete(r.ID))
			assert.Must(t, Tracker.Delete(tracker.ID))
			assert.Must(t, Identity.Delete(identity.ID))
			assert.Must(t, Application.Delete(app.ID))

			// Check if the Ticket is present even after deletion or not.
			_, err = Ticket.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestTicketList(t *testing.T) {
	createdTickets := []api.Ticket{}

	for _, r := range Samples {
		// Create a sample Application for the ticket.
		app := api.Application{
			Name: r.Application.Name,
		}
		assert.Must(t, Application.Create(&app))

		// Create a sample identity for the tracker
		identity := api.Identity{
			Name: "Sample Identity",
			Kind: "Sample Kind",
		}
		assert.Must(t, Identity.Create(&identity))

		// create a sample tracker for the ticket.
		tracker := api.Tracker{
			Name: r.Tracker.Name,
			URL:  "www.sample.com",
			Kind: "jira-onprem",
			Identity: api.Ref{
				ID:   1,
				Name: "Sample Identity",
			},
		}
		assert.Must(t, Tracker.Create(&tracker))

		// Create a sample ticket
		assert.Must(t, Ticket.Create(&r))
		createdTickets = append(createdTickets, r)
	}

	// List Tickets.
	got, err := Ticket.List()
	if err != nil {
		t.Errorf(err.Error())
	}

	// check if created Tickets are in the list we got from Ticket.List()
	for _, createdTicket := range createdTickets {
		found := false
		for _, retrievedTicket := range got {
			if assert.FlatEqual(createdTicket.ID, retrievedTicket.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tracker not found in the list: %v", createdTicket)
		}
	}

	// Delete tickets and related resources.
	for _, ticket := range createdTickets {
		assert.Must(t, Ticket.Delete(ticket.ID))
		assert.Must(t, Tracker.Delete(ticket.ID))
		assert.Must(t, Identity.Delete(ticket.ID))
		assert.Must(t, Application.Delete(ticket.ID))
	}
}

func AssertEqualTickets(t *testing.T, got *api.Ticket, expected api.Ticket) {
	if got.Kind != expected.Kind {
		t.Errorf("Different Kind Got %v, expected %v", got.Kind, expected.Kind)
	}
	if got.Reference != expected.Reference {
		t.Errorf("Different Tracker Reference Got %v, expected %v", got.Reference, expected.Reference)
	}
	if got.Link != expected.Link {
		t.Errorf("Different Url Got %v, expected %v", got.Link, expected.Link)
	}
	if got.Parent != expected.Parent {
		t.Errorf("Different Parent Got %v, expected %v", got.Parent, expected.Parent)
	}
	if got.Message != expected.Message {
		t.Errorf("Different Message Got %v, expected %v", got.Message, expected.Message)
	}
	if got.Status != expected.Status {
		t.Errorf("Different Status Got %v, expected %v", got.Status, expected.Status)
	}
	if got.Application.Name != expected.Application.Name {
		t.Errorf("Different Application's Name Got %v, expected %v", got.Application.Name, expected.Application.Name)
	}
	if got.Tracker.Name != expected.Tracker.Name {
		t.Errorf("Different Tracker's Name Got %v, expected %v", got.Tracker.Name, expected.Tracker.Name)
	}
}
