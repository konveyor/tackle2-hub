package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/konveyor/tackle2-hub/internal/api/filter"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/api/sort"
	"github.com/konveyor/tackle2-hub/internal/jsd"
	"github.com/konveyor/tackle2-hub/internal/model"
	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// BadRequestError reports bad request errors.
type BadRequestError struct {
	Reason string
}

func (r *BadRequestError) Error() string {
	return r.Reason
}

func (r *BadRequestError) Is(err error) (matched bool) {
	_, matched = err.(*BadRequestError)
	return
}

// BatchError reports errors stemming from batch operations.
type BatchError struct {
	Message string
	Items   []BatchErrorItem
}

type BatchErrorItem struct {
	Error    error
	Resource any
}

func (r BatchError) Error() string {
	return r.Message
}

func (r BatchError) Is(err error) (matched bool) {
	_, matched = err.(BatchError)
	return
}

// TrackerError reports an error stemming from the Hub being unable
// to communicate with an external issue tracker.
type TrackerError struct {
	Reason string
}

func (r *TrackerError) Error() string {
	return r.Reason
}

func (r *TrackerError) Is(err error) (matched bool) {
	_, matched = err.(*TrackerError)
	return
}

// Forbidden reports auth errors.
type Forbidden struct {
	Reason string
}

func (r *Forbidden) Error() string {
	return r.Reason
}

func (r *Forbidden) Is(err error) (matched bool) {
	_, matched = err.(*Forbidden)
	return
}

// NotFound reports resource not-found errors.
type NotFound struct {
	Resource string
	Reason   string
}

func (r *NotFound) Error() string {
	return fmt.Sprintf("Resource '%s' not found. %s", r.Resource, r.Reason)
}

func (r *NotFound) Is(err error) (matched bool) {
	var forbidden *Forbidden
	matched = errors.As(err, &forbidden)
	return
}

// ErrorHandler handles error conditions from lower handlers.
func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors[0]

		rtx := RichContext(ctx)
		if errors.Is(err, &BadRequestError{}) ||
			errors.Is(err, &resource.ValidationError{}) ||
			errors.Is(err, &filter.Error{}) ||
			errors.Is(err, &sort.SortError{}) ||
			errors.Is(err, validator.ValidationErrors{}) {
			rtx.Respond(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) ||
			errors.Is(err, &NotFound{}) ||
			errors.Is(err, &jsd.NotFound{}) {
			if ctx.Request.Method == http.MethodDelete {
				rtx.Status(http.StatusNoContent)
				return
			}
			rtx.Respond(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, os.ErrNotExist) {
			rtx.Respond(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, model.DependencyCyclicError{}) {
			rtx.Respond(
				http.StatusConflict,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, &TrackerError{}) {
			rtx.Respond(
				http.StatusServiceUnavailable,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, &Forbidden{}) {
			rtx.Respond(
				http.StatusForbidden,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		sqliteErr := &sqlite3.Error{}
		if errors.As(err, sqliteErr) {
			switch sqliteErr.ExtendedCode {
			case sqlite3.ErrConstraintUnique,
				sqlite3.ErrConstraintPrimaryKey:
				rtx.Respond(
					http.StatusConflict,
					gin.H{
						"error": err.Error(),
					})
				return
			}
		}

		bErr := &BatchError{}
		if errors.As(err, bErr) {
			rtx.Respond(
				http.StatusBadRequest,
				bErr.Items,
			)
			return
		}

		if errors.Is(err, &tasking.BadRequest{}) {
			rtx.Respond(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		rtx.Respond(
			http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			})

		url := ctx.Request.URL.String()
		log.Error(
			err.Err,
			"Request failed.",
			"method",
			ctx.Request.Method,
			"url",
			url)
	}
}
