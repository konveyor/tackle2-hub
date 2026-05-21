package auth

import (
	"errors"
	"fmt"

	"github.com/konveyor/tackle2-hub/internal/auth/cache"
)

// BadRequestError reports bad request errors.
type BadRequestError struct {
	Reason string
}

// Error returns the error message.
func (r *BadRequestError) Error() string {
	return r.Reason
}

// Is returns true if the error is a BadRequestError.
func (r *BadRequestError) Is(err error) (matched bool) {
	var target *BadRequestError
	matched = errors.As(err, &target)
	return
}

type ScopeNotFound struct {
	Scope string
}

func (e *ScopeNotFound) Error() string {
	return fmt.Sprintf("Scope %s not-found.", e.Scope)
}

func (e *ScopeNotFound) Is(err error) (matched bool) {
	var inst *ScopeNotFound
	matched = errors.As(err, &inst)
	return
}

// NotFound reports not found.
type NotFound = cache.NotFound
