package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Ticket REST resource.
type Ticket api.Ticket

// With updates the resource with the model.
func (r *Ticket) With(m *model.Ticket) {
	baseWith(&r.Resource, &m.Model)
	r.Kind = m.Kind
	r.Reference = m.Reference
	r.Parent = m.Parent
	r.Link = m.Link
	r.Error = m.Error
	r.Message = m.Message
	r.Status = m.Status
	r.LastUpdated = m.LastUpdated
	r.Application = ref(m.ApplicationID, m.Application)
	r.Tracker = ref(m.TrackerID, m.Tracker)
	r.Fields = m.Fields
}

// Model builds a model.
func (r *Ticket) Model() (m *model.Ticket) {
	m = &model.Ticket{
		Kind:          r.Kind,
		Parent:        r.Parent,
		ApplicationID: r.Application.ID,
		TrackerID:     r.Tracker.ID,
	}
	m.Fields = r.Fields
	m.ID = r.ID

	return
}
