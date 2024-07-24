package reflect

import (
	"errors"
	"fmt"
)

// FieldNotValid report field not valid.
type FieldNotValid struct {
	Kind string
	Name string
}

func (e *FieldNotValid) Error() string {
	return fmt.Sprintf(
		"(%s) '%s' not valid.",
		e.Kind,
		e.Name)
}

func (e *FieldNotValid) Is(err error) (matched bool) {
	var inst *FieldNotValid
	matched = errors.As(err, &inst)
	return
}
