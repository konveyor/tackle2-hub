package resource

import (
	"time"

	"github.com/konveyor/tackle2-hub/model"
)

// Tracker REST resource.
type Tracker struct {
	Resource    `yaml:",inline"`
	Name        string    `json:"name" binding:"required"`
	URL         string    `json:"url" binding:"required"`
	Kind        string    `json:"kind" binding:"required,oneof=jira-cloud jira-onprem"`
	Message     string    `json:"message"`
	Connected   bool      `json:"connected"`
	LastUpdated time.Time `json:"lastUpdated" yaml:"lastUpdated"`
	Identity    Ref       `json:"identity" binding:"required"`
	Insecure    bool      `json:"insecure"`
}

// With updates the resource with the model.
func (r *Tracker) With(m *model.Tracker) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.URL = m.URL
	r.Kind = m.Kind
	r.Message = m.Message
	r.Connected = m.Connected
	r.LastUpdated = m.LastUpdated
	r.Insecure = m.Insecure
	r.Identity = r.ref(m.IdentityID, m.Identity)
}

// Model builds a model.
func (r *Tracker) Model() (m *model.Tracker) {
	m = &model.Tracker{
		Name:       r.Name,
		URL:        r.URL,
		Kind:       r.Kind,
		Insecure:   r.Insecure,
		IdentityID: r.Identity.ID,
	}

	m.ID = r.ID

	return
}
