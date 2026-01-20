package binding

import (
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

const (
	// RetryLimit request retry limit.
	RetryLimit = client.RetryLimit

	// RetryDelay delay between client request retries.
	RetryDelay = client.RetryDelay
)

// Filter used to filter List()
type Filter = client.Filter

// Param API parameter.
type Param = client.Param

// Params API parameters.
type Params = client.Params

// Path URL path.
type Path = client.Path

// Field Http form field.
type Field = client.Field

// Client Http client.
type Client = client.Client

// RestError reports REST errors.
type RestError = client.RestError

// Conflict reports 409 error.
type Conflict = client.Conflict

// NotFound reports 404 error.
type NotFound = client.NotFound

// EmptyBody reports an empty body.
type EmptyBody = client.EmptyBody
