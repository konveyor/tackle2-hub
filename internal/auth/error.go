package auth

import "errors"

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
