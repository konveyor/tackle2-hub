package resource

// ValidationError REST resource.
type ValidationError struct {
	Reason string
}

func (r *ValidationError) Error() string {
	return r.Reason
}

func (r *ValidationError) Is(err error) (matched bool) {
	_, matched = err.(*ValidationError)
	return
}
