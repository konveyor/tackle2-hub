package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Tracker REST resource.
type Tracker api.Tracker

// With updates the resource with the model.
func (r *Tracker) With(m *model.Tracker) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.URL = m.URL
	r.Kind = m.Kind
	r.Message = m.Message
	r.Connected = m.Connected
	r.LastUpdated = m.LastUpdated
	r.Insecure = m.Insecure
	r.Identity = ref(m.IdentityID, m.Identity)
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

// IssueType REST resource.
type IssueType = api.IssueType

// Project REST resource.
type Project = api.Project
