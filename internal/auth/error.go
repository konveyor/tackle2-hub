package auth

import (
	"errors"
	"fmt"

	"github.com/konveyor/tackle2-hub/internal/auth/cache"
)

// NotAuthenticated is returned when a token cannot be authenticated.
type NotAuthenticated struct {
	Token  string
	Reason string
}

func (e *NotAuthenticated) Error() (s string) {
	if e.Reason != "" {
		return fmt.Sprintf("Token [%s] not-authenticated: %s", e.Token, e.Reason)
	}
	return fmt.Sprintf("Token [%s] not-authenticated.", e.Token)
}

func (e *NotAuthenticated) Is(err error) (matched bool) {
	notAuth := &NotAuthenticated{}
	matched = errors.As(err, &notAuth)
	return
}

// NotValid is returned when a token is not valid.
type NotValid struct {
	Reason  string
	TokenId string
}

func (e *NotValid) Error() (s string) {
	return fmt.Sprintf("Token (jti=%s) not-valid: %s", e.TokenId, e.Reason)
}

func (e *NotValid) Is(err error) (matched bool) {
	notValid := &NotValid{}
	matched = errors.As(err, &notValid)
	return
}

// BadRequestError reports bad request errors.
type BadRequestError struct {
	Reason string
}

// Error returns the error message.
func (r *BadRequestError) Error() string {
	return r.Reason
}

// Is returns true if the error is a BadRequestError.
func (r *BadRequestError) Is(err error) (matched bool) {
	var target *BadRequestError
	matched = errors.As(err, &target)
	return
}

type ScopeNotFound struct {
	Scope string
}

func (e *ScopeNotFound) Error() string {
	return fmt.Sprintf("Scope %s not-found.", e.Scope)
}

func (e *ScopeNotFound) Is(err error) (matched bool) {
	var inst *ScopeNotFound
	matched = errors.As(err, &inst)
	return
}

// NotFound reports not found.
type NotFound = cache.NotFound
