package task

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
)

// BadRequest report bad request.
type BadRequest struct {
	Reason string
}

func (e *BadRequest) Error() string {
	return e.Reason
}

func (e *BadRequest) Is(err error) (matched bool) {
	var inst *BadRequest
	matched = errors.As(err, &inst)
	return
}

// SoftErr returns true when the error isA SoftError.
func SoftErr(err error) (matched, retry bool) {
	if err == nil {
		return
	}
	naked := errors.Unwrap(err)
	if naked == nil {
		naked = err
	}
	if softErr, cast := naked.(SoftError); cast {
		matched = true
		retry = softErr.Retry()
	}
	return
}

// SoftError used to report errors specific to one task
// rather than systemic issues.
type SoftError interface {
	// Retry determines if the task should be
	// retried or failed immediately.
	Retry() (r bool)
}

// KindNotFound used to report task (kind) referenced
// by a task but cannot be found.
type KindNotFound struct {
	Name string
}

func (e *KindNotFound) Error() string {
	return fmt.Sprintf(
		"Task (kind): '%s' not-found.",
		e.Name)
}

func (e *KindNotFound) Is(err error) (matched bool) {
	var inst *KindNotFound
	matched = errors.As(err, &inst)
	return
}

func (e *KindNotFound) Retry() (r bool) {
	return
}

// AddonNotFound used to report addon referenced
// by a task but cannot be found.
type AddonNotFound struct {
	Name string
}

func (e *AddonNotFound) Error() string {
	return fmt.Sprintf(
		"Addon: '%s' not-found.",
		e.Name)
}

func (e *AddonNotFound) Is(err error) (matched bool) {
	var inst *AddonNotFound
	matched = errors.As(err, &inst)
	return
}

func (e *AddonNotFound) Retry() (r bool) {
	return
}

// AddonNotSelected report that an addon has not been selected.
type AddonNotSelected struct {
}

func (e *AddonNotSelected) Error() string {
	return fmt.Sprintf("Addon not selected.")
}

func (e *AddonNotSelected) Is(err error) (matched bool) {
	var inst *AddonNotSelected
	matched = errors.As(err, &inst)
	return
}

func (e *AddonNotSelected) Retry() (r bool) {
	return
}

// NotReady report that a resource does not have the ready condition.
type NotReady struct {
	Kind   string
	Name   string
	Reason string
}

func (e *NotReady) Error() string {
	return fmt.Sprintf(
		"(%s) '%s' not ready: %s.",
		e.Kind,
		e.Name,
		e.Reason)
}

func (e *NotReady) Is(err error) (matched bool) {
	var inst *NotReady
	matched = errors.As(err, &inst)
	return
}

func (e *NotReady) Retry() (r bool) {
	return
}

// NotReconciled report as resource has not been reconciled.
type NotReconciled struct {
	Kind string
	Name string
}

func (e *NotReconciled) Error() string {
	return fmt.Sprintf("(%s) '%s' not reconciled.", e.Kind, e.Name)
}

func (e *NotReconciled) Is(err error) (matched bool) {
	var inst *NotReconciled
	matched = errors.As(err, &inst)
	return
}

func (e *NotReconciled) Retry() (r bool) {
	return
}

// ExtensionNotFound used to report an extension referenced
// by a task but cannot be found.
type ExtensionNotFound struct {
	Name string
}

func (e *ExtensionNotFound) Error() string {
	return fmt.Sprintf(
		"Extension: '%s' not-found.",
		e.Name)
}

func (e *ExtensionNotFound) Is(err error) (matched bool) {
	var inst *ExtensionNotFound
	matched = errors.As(err, &inst)
	return
}

func (e *ExtensionNotFound) Retry() (r bool) {
	return
}

// ExtensionNotValid used to report extension referenced
// by a task not valid with addon.
type ExtensionNotValid struct {
	Name  string
	Addon string
}

func (e *ExtensionNotValid) Error() string {
	return fmt.Sprintf(
		"Extension: '%s' not-valid with addon '%s'.",
		e.Name,
		e.Addon)
}

func (e *ExtensionNotValid) Is(err error) (matched bool) {
	var inst *ExtensionNotValid
	matched = errors.As(err, &inst)
	return
}

func (e *ExtensionNotValid) Retry() (r bool) {
	return
}

// SelectorNotValid reports selector errors.
type SelectorNotValid struct {
	Selector  string
	Predicate string
	Reason    string
}

func (e *SelectorNotValid) Error() string {
	if e.Predicate != "" {
		return fmt.Sprintf(
			"Selector '%s' not valid. predicate '%s' not supported.",
			e.Selector,
			e.Predicate)
	}
	return fmt.Sprintf(
		"Selector syntax '%s' not valid: '%s'.",
		e.Selector,
		e.Reason)
}

func (e *SelectorNotValid) Is(err error) (matched bool) {
	var inst *SelectorNotValid
	matched = errors.As(err, &inst)
	return
}

func (e *SelectorNotValid) Retry() (r bool) {
	return
}

// ExtAddonNotValid reports extension addon ref error.
type ExtAddonNotValid struct {
	Extension string
	Reason    string
}

func (e *ExtAddonNotValid) Error() string {
	return fmt.Sprintf(
		"Extension '%s' addon ref not valid. reason: %s",
		e.Extension,
		e.Reason)
}

func (e *ExtAddonNotValid) Is(err error) (matched bool) {
	var inst *ExtAddonNotValid
	matched = errors.As(err, &inst)
	return
}

// AddonTaskNotValid reports addon task ref error.
type AddonTaskNotValid struct {
	Addon  string
	Reason string
}

func (e *AddonTaskNotValid) Error() string {
	return fmt.Sprintf(
		"Addon '%s' task ref not valid. reason: %s",
		e.Addon,
		e.Reason)
}

func (e *AddonTaskNotValid) Is(err error) (matched bool) {
	var inst *AddonTaskNotValid
	matched = errors.As(err, &inst)
	return
}

// PriorityNotFound report priority class not found.
type PriorityNotFound struct {
	Name  string
	Value int
}

func (e *PriorityNotFound) Error() string {
	var d string
	if e.Name != "" {
		d = fmt.Sprintf("\"%s\"", e.Name)
	} else {
		d = strconv.Itoa(e.Value)
	}
	return fmt.Sprintf(
		"PriorityClass %s not-found.",
		d)
}

func (e *PriorityNotFound) Is(err error) (matched bool) {
	var inst *PriorityNotFound
	matched = errors.As(err, &inst)
	return
}

func (e *PriorityNotFound) Retry() (r bool) {
	return
}

// PodRejected report pod rejected..
type PodRejected struct {
	Reason string
}

func (e *PodRejected) Error() string {
	return e.Reason
}

func (e *PodRejected) Is(err error) (matched bool) {
	var inst *PodRejected
	matched = errors.As(err, &inst)
	return
}

// Match returns true when pod is rejected.
func (e *PodRejected) Match(err error) (matched bool) {
	matched = k8serr.IsBadRequest(err) ||
		k8serr.IsForbidden(err) ||
		k8serr.IsInvalid(err)
	if matched {
		e.Reason = err.Error()
	}
	return
}

func (e *PodRejected) Retry() (r bool) {
	return
}

// QuotaExceeded report quota exceeded.
type QuotaExceeded struct {
	Reason string
}

// Match returns true when the error is Forbidden due to quota exceeded.
func (e *QuotaExceeded) Match(err error) (matched bool) {
	if k8serr.IsForbidden(err) {
		matched = true
		e.Reason = err.Error()
		for _, s := range []string{"quota", "exceeded"} {
			matched = strings.Contains(e.Reason, s)
			if !matched {
				break
			}
		}
		part := strings.SplitN(e.Reason, ":", 2)
		if len(part) > 1 {
			e.Reason = part[1]
		}
	}
	return
}

func (e *QuotaExceeded) Error() string {
	return e.Reason
}

func (e *QuotaExceeded) Is(err error) (matched bool) {
	var inst *QuotaExceeded
	matched = errors.As(err, &inst)
	return
}

func (e *QuotaExceeded) Retry() (r bool) {
	r = true
	return
}
