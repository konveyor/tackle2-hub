package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Ticket API.
type Ticket struct {
	client *Client
}

// Create a Ticket.
func (h *Ticket) Create(r *api.Ticket) (err error) {
	err = h.client.Post(api.TicketsRoute, &r)
	return
}

// Get a Ticket by ID.
func (h *Ticket) Get(id uint) (r *api.Ticket, err error) {
	r = &api.Ticket{}
	path := Path(api.TicketRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tickets..
func (h *Ticket) List() (list []api.Ticket, err error) {
	list = []api.Ticket{}
	err = h.client.Get(api.TicketsRoute, &list)
	return
}

// Delete a Ticket.
func (h *Ticket) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TicketRoute).Inject(Params{api.ID: id}))
	return
}
