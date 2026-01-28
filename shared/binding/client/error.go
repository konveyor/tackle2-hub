package client

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// RestError reports REST errors.
type RestError = api.RestError

// BadRequestError reports bad request errors.
type BadRequestError = api.BadRequestError

// Forbidden reports auth errors.
type Forbidden = api.Forbidden

// Conflict reports 409 error.
type Conflict = api.Conflict

// NotFound reports 404 error.
type NotFound = api.NotFound

// EmptyBody reports an empty body.
type EmptyBody = api.EmptyBody
