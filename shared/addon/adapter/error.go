package adapter

// SoftError A "soft" anticipated error.
// Deprecated:
type SoftError struct {
	Reason string
}

func (e *SoftError) Error() (s string) {
	return e.Reason
}

func (e *SoftError) Is(err error) (matched bool) {
	_, matched = err.(*SoftError)
	return
}

func (e *SoftError) Soft() *SoftError {
	return e
}
