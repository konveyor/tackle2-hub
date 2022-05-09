package addon

import (
	"fmt"
)

//
// SoftError A "soft" anticipated error.
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

//
// Conflict reports 409 error.
type Conflict struct {
	SoftError
	Path string
}

func (e Conflict) Error() string {
	return fmt.Sprintf("POST: path:%s (conflict)", e.Path)
}

func (e *Conflict) Is(err error) (matched bool) {
	_, matched = err.(*Conflict)
	return
}

//
// NotFound reports 404 error.
type NotFound struct {
	SoftError
	Path string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("HTTP path:%s (not-found)", e.Path)
}

func (e *NotFound) Is(err error) (matched bool) {
	_, matched = err.(*NotFound)
	return
}
