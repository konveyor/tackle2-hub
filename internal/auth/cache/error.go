package cache

import (
	"errors"
	"fmt"
)

// NotFound reports not found.
type NotFound struct {
	Resource string
	Filter   string
	Id       string
}

// Error returns the error message.
func (r *NotFound) Error() (s string) {
	s = fmt.Sprintf("%s (%s) not-found", r.Resource, r.Id)
	if r.Filter != "" {
		s += ", filter: " + r.Filter
	}
	return
}

// Is returns true if the error is a NotFound.
func (r *NotFound) Is(err error) (matched bool) {
	var target *NotFound
	matched = errors.As(err, &target)
	return
}
