package cache

import (
	"errors"
	"fmt"
)

// NotFound reports not found.
type NotFound struct {
	Resource string
	Id       string
}

// Error returns the error message.
func (r *NotFound) Error() string {
	return fmt.Sprintf("%s (%s) not-found", r.Resource, r.Id)
}

// Is returns true if the error is a NotFound.
func (r *NotFound) Is(err error) (matched bool) {
	var target *NotFound
	matched = errors.As(err, &target)
	return
}
