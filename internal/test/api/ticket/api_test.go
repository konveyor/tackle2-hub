package ticket

import (
	"testing"

	api2 "github.com/konveyor/tackle2-hub/api"
	TrackerSamples "github.com/konveyor/tackle2-hub/internal/test/api/tracker"
	assert2 "github.com/konveyor/tackle2-hub/internal/test/assert"
)

func TestTicketCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("Ticket "+r.Kind+" CRUD", func(t *testing.T) {

			// Create a sample Application for the ticket.
			app := api2.Application{
				Name: r.Application.Name,
			}
			assert2.Must(t, Application.Create(&app))
			r.Application.ID = app.ID

			createdIdentities := []api2.Identity{}
			createdTrackers := []api2.Tracker{}
			for _, tracker := range TrackerSamples.Samples {
				// Create a sample identity for the tracker
				identity := api2.Identity{
					Name: tracker.Identity.Name,
					Kind: tracker.Kind,
				}
				assert2.Must(t, Identity.Create(&identity))
				tracker.Identity.ID = identity.ID
				createdIdentities = append(createdIdentities, identity)
				assert2.Must(t, Tracker.Create(&tracker))
				r.Tracker.ID = tracker.ID
				r.Tracker.Name = tracker.Name
				createdTrackers = append(createdTrackers, tracker)
			}

			// Create a sample ticket
			assert2.Must(t, Ticket.Create(&r))

			// Get.
			got, err := Ticket.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare got values with expected values.
			AssertEqualTickets(t, got, r)

			// Delete ticket and its related resources.
			assert2.Must(t, Ticket.Delete(r.ID))
			for _, tracker := range createdTrackers {
				assert2.Must(t, Tracker.Delete(tracker.ID))
			}
			for _, identity := range createdIdentities {
				assert2.Must(t, Identity.Delete(identity.ID))
			}
			assert2.Must(t, Application.Delete(app.ID))

			// Check if the Ticket is present even after deletion or not.
			_, err = Ticket.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestTicketList(t *testing.T) {
	for _, r := range Samples {

		createdTickets := []api2.Ticket{}
		// Create a sample Application for the ticket.
		app := api2.Application{
			Name: r.Application.Name,
		}
		assert2.Must(t, Application.Create(&app))
		r.Application.ID = app.ID

		createdIdentities := []api2.Identity{}
		createdTrackers := []api2.Tracker{}
		for _, tracker := range TrackerSamples.Samples {
			// Create a sample identity for the tracker
			identity := api2.Identity{
				Name: tracker.Identity.Name,
				Kind: tracker.Kind,
			}
			assert2.Must(t, Identity.Create(&identity))
			tracker.Identity.ID = identity.ID
			createdIdentities = append(createdIdentities, identity)
			assert2.Must(t, Tracker.Create(&tracker))
			r.Tracker.ID = tracker.ID
			r.Tracker.Name = tracker.Name
			createdTrackers = append(createdTrackers, tracker)
		}

		// Create a sample ticket
		assert2.Must(t, Ticket.Create(&r))
		createdTickets = append(createdTickets, r)

		// List Tickets.
		got, err := Ticket.List()
		if err != nil {
			t.Errorf(err.Error())
		}

		for _, createdTicket := range createdTickets {
			found := false
			for _, retrievedTicket := range got {
				if assert2.FlatEqual(createdTicket.ID, retrievedTicket.ID) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected ticket not found in the list: %v", createdTicket)
			}
		}

		// Delete tickets and related resources.
		for _, ticket := range createdTickets {
			assert2.Must(t, Ticket.Delete(ticket.ID))
			assert2.Must(t, Application.Delete(ticket.Application.ID))
		}
		for _, tracker := range createdTrackers {
			assert2.Must(t, Tracker.Delete(tracker.ID))
		}
		for _, identity := range createdIdentities {
			assert2.Must(t, Identity.Delete(identity.ID))
		}
	}
}

func AssertEqualTickets(t *testing.T, got *api2.Ticket, expected api2.Ticket) {
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
