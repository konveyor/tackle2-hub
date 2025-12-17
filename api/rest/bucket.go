package rest

import (
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Bucket type alias to shared API.
type Bucket api.Bucket

// With updates the resource with the model.
func (r *Bucket) With(m *model.Bucket) {
	baseWith(&r.Resource, &m.Model)
	r.Path = m.Path
	r.Expiration = m.Expiration
}
