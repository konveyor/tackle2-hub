package resource

import (
	"time"

	"github.com/konveyor/tackle2-hub/model"
)

// Ticket API Resource
type Ticket struct {
	Resource    `yaml:",inline"`
	Kind        string    `json:"kind" binding:"required"`
	Reference   string    `json:"reference"`
	Link        string    `json:"link"`
	Parent      string    `json:"parent" binding:"required"`
	Error       bool      `json:"error"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated" yaml:"lastUpdated"`
	Fields      Map       `json:"fields"`
	Application Ref       `json:"application" binding:"required"`
	Tracker     Ref       `json:"tracker" binding:"required"`
}

// With updates the resource with the model.
func (r *Ticket) With(m *model.Ticket) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Reference = m.Reference
	r.Parent = m.Parent
	r.Link = m.Link
	r.Error = m.Error
	r.Message = m.Message
	r.Status = m.Status
	r.LastUpdated = m.LastUpdated
	r.Application = r.ref(m.ApplicationID, m.Application)
	r.Tracker = r.ref(m.TrackerID, m.Tracker)
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
