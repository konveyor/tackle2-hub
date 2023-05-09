package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"net/http"
	"os"
)

//
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

//
// BatchError reports errors stemming from batch operations.
type BatchError struct {
	Message string
	Items   []BatchErrorItem
}

type BatchErrorItem struct {
	Error    error
	Resource interface{}
}

func (r BatchError) Error() string {
	return r.Message
}

func (r BatchError) Is(err error) (matched bool) {
	_, matched = err.(BatchError)
	return
}

//
// ErrorHandler handles error conditions from lower handlers.
func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		err := ctx.Errors[0]

		rtx := WithContext(ctx)
		if errors.Is(err, &BadRequestError{}) ||
			errors.Is(err, &filter.Error{}) ||
			errors.Is(err, validator.ValidationErrors{}) {
			rtx.Respond(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
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
