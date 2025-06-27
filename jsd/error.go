package jsd

import "errors"

type NotFound struct {
}

func (e *NotFound) Error() string {
	return "Not Found"
}

func (e *NotFound) Is(err error) (matched bool) {
	var inst *NotFound
	matched = errors.As(err, &inst)
	return
}

type NotValid struct {
	Reason string
}

func (e *NotValid) Error() string {
	return "Not Valid: " + e.Reason
}

func (e *NotValid) Is(err error) (matched bool) {
	var inst *NotValid
	matched = errors.As(err, &inst)
	return
}
