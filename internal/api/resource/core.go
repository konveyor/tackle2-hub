package resource

import (
	"time"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// LatestSchema REST resource.
type LatestSchema = api.LatestSchema

// Cache REST resource.
type Cache = api.Cache

// ConfigMap REST resource.
type ConfigMap = api.ConfigMap

// APIKey REST resource.
type APIKey api.APIKey

// With converts model to REST resource.
func (r *APIKey) With(m *model.APIKey) {
	baseWith(&r.Resource, &m.Model)
	r.Digest = m.Digest
	r.Lifespan = int(time.Until(m.Expiration) / time.Hour)
	r.Expired = time.Now().After(m.Expiration)
	r.Expiration = m.Expiration
	if m.User != nil {
		r.User = &api.Ref{
			ID:   m.User.ID,
			Name: m.User.Userid,
		}
	}
	if m.Task != nil {
		r.Task = &api.Ref{
			ID:   m.Task.ID,
			Name: m.Task.Name,
		}
	}
}

// RestAPI REST resource.
type RestAPI = api.RestAPI

// Service REST resource.
type Service = api.Service
