package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Bucket REST resource.
type Bucket api.Bucket

// With updates the resource with the model.
func (r *Bucket) With(m *model.Bucket) {
	baseWith(&r.Resource, &m.Model)
	r.Path = m.Path
	r.Expiration = m.Expiration
}
