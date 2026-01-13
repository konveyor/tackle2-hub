package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Ticket API.
type Ticket struct {
	client *Client
}

// Create a Ticket.
func (h *Ticket) Create(r *api2.Ticket) (err error) {
	err = h.client.Post(api2.TicketsRoute, &r)
	return
}

// Get a Ticket by ID.
func (h *Ticket) Get(id uint) (r *api2.Ticket, err error) {
	r = &api2.Ticket{}
	path := Path(api2.TicketRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Tickets..
func (h *Ticket) List() (list []api2.Ticket, err error) {
	list = []api2.Ticket{}
	err = h.client.Get(api2.TicketsRoute, &list)
	return
}

// Delete a Ticket.
func (h *Ticket) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.TicketRoute).Inject(Params{api2.ID: id}))
	return
}
