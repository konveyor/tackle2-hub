package api

import (
	"errors"
	"fmt"
)

// BadRequestError reports bad request errors.
type BadRequestError struct {
	Reason string
}

func (r *BadRequestError) Error() string {
	return r.Reason
}

func (r *BadRequestError) Is(err error) (matched bool) {
	var target *BadRequestError
	matched = errors.As(err, &target)
	return
}

// Forbidden reports auth errors.
type Forbidden struct {
	Reason string
}

func (r *Forbidden) Error() string {
	return r.Reason
}

func (r *Forbidden) Is(err error) (matched bool) {
	var target *Forbidden
	matched = errors.As(err, &target)
	return
}

// NotFound reports resource not-found errors.
type NotFound struct {
	Resource string
	Reason   string
}

func (r *NotFound) Error() string {
	return fmt.Sprintf("Resource '%s' not found. %s", r.Resource, r.Reason)
}

func (r *NotFound) Is(err error) (matched bool) {
	var target *NotFound
	matched = errors.As(err, &target)
	return
}
