package jsd

type NotFound struct {
}

func (e *NotFound) Error() string {
	return "Not Found"
}

func (e *NotFound) Is(err error) (matched bool) {
	_, matched = err.(*NotFound)
	return
}

type NotValid struct {
	Reason string
}

func (e *NotValid) Error() string {
	return "Not Valid: " + e.Reason
}

func (e *NotValid) Is(err error) (matched bool) {
	_, matched = err.(*NotValid)
	return
}
