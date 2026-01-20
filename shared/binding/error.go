package binding

import (
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// RestError reports REST errors.
type RestError = client.RestError

// Conflict reports 409 error.
type Conflict = client.Conflict

// NotFound reports 404 error.
type NotFound = client.NotFound

// EmptyBody reports an empty body.
type EmptyBody = client.EmptyBody
