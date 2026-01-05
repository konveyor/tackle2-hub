package filter

import "fmt"

// Error reports bad request errors.
type Error struct {
	Reason string
}

func (r *Error) Error() string {
	return "filter: " + r.Reason
}

func (r *Error) Is(err error) (matched bool) {
	_, matched = err.(*Error)
	return
}

// Errorf build error.
func Errorf(s string, v ...any) (err error) {
	err = &Error{fmt.Sprintf(s, v...)}
	return
}
