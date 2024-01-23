package sort

import "fmt"

// SortError reports sorting error.
type SortError struct {
	field string
}

func (r *SortError) Error() string {
	return fmt.Sprintf("\"%s\" not supported by sort.", r.field)
}

func (r *SortError) Is(err error) (matched bool) {
	_, matched = err.(*SortError)
	return
}
