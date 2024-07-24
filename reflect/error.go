package reflect

import (
	"errors"
	"fmt"
)

// FieldNotKnown report field not found.
type FieldNotKnown struct {
	Kind string
	Name string
}

func (e *FieldNotKnown) Error() string {
	return fmt.Sprintf(
		"(%s) '%s' not known.",
		e.Kind,
		e.Name)
}

func (e *FieldNotKnown) Is(err error) (matched bool) {
	var inst *FieldNotKnown
	matched = errors.As(err, &inst)
	return
}
