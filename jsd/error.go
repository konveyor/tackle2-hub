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
