package resource

import (
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Scope REST resource.
type Scope api.Scope

// With converts string to REST resource.
func (r *Scope) With(m string) {
	scope := auth.Scope{}
	scope.With(m)
	r.Name = m
	r.Resource = scope.Resource
	r.Verb = scope.Method
}
