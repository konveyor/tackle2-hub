package task

import (
	"errors"
	"fmt"
)

// KindNotFound used to report profile referenced
// by a task but cannot be found.
type KindNotFound struct {
	Name string
}

func (e *KindNotFound) Error() (s string) {
	return fmt.Sprintf("Task (kind): '%s' not-found.", e.Name)
}

func (e *KindNotFound) Is(err error) (matched bool) {
	var inst *KindNotFound
	matched = errors.As(err, &inst)
	return
}

// AddonNotFound used to report addon referenced
// by a task but cannot be found.
type AddonNotFound struct {
	Name string
}

func (e *AddonNotFound) Error() (s string) {
	return fmt.Sprintf("Addon: '%s' not-found.", e.Name)
}

func (e *AddonNotFound) Is(err error) (matched bool) {
	var inst *AddonNotFound
	matched = errors.As(err, &inst)
	return
}

// AddonNotSelected report that an addon has not been selected.
type AddonNotSelected struct {
}

func (e *AddonNotSelected) Error() (s string) {
	return fmt.Sprintf("Addon not selected.")
}

func (e *AddonNotSelected) Is(err error) (matched bool) {
	var inst *AddonNotSelected
	matched = errors.As(err, &inst)
	return
}

// ExtensionNotFound used to report extension referenced
// by a task but cannot be found.
type ExtensionNotFound struct {
	Name string
}

func (e *ExtensionNotFound) Error() (s string) {
	return fmt.Sprintf("Extension: '%s' not-found.", e.Name)
}

func (e *ExtensionNotFound) Is(err error) (matched bool) {
	var inst *ExtensionNotFound
	matched = errors.As(err, &inst)
	return
}

// ExtensionNotValid used to report extension referenced
// by a task not valid with addon.
type ExtensionNotValid struct {
	Name string
}

func (e *ExtensionNotValid) Error() (s string) {
	return fmt.Sprintf("Extension: '%s' not-valid with addon.", e.Name)
}

func (e *ExtensionNotValid) Is(err error) (matched bool) {
	var inst *ExtensionNotValid
	matched = errors.As(err, &inst)
	return
}

// UnknownSelector reports unknown selector.
type UnknownSelector struct {
	Kind string
}

func (e *UnknownSelector) Error() (s string) {
	return fmt.Sprintf("Selector: '%s' unknown. Not supported.", e.Kind)
}

func (e *UnknownSelector) Is(err error) (matched bool) {
	var inst *UnknownSelector
	matched = errors.As(err, &inst)
	return
}

// NotResolved report name/capability not resolved.
type NotResolved struct {
	Kind string
	Name string
}

func (e *NotResolved) Error() (s string) {
	return fmt.Sprintf("%s: '%s' not-resolved.", e.Kind, e.Name)
}

func (e *NotResolved) Is(err error) (matched bool) {
	var inst *NotResolved
	matched = errors.As(err, &inst)
	return
}
