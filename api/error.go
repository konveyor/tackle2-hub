package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 {
			return
		}

		if len(ctx.Errors) > 1 {
			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"errors": ctx.Errors.Errors(),
				})
			return
		}

		err := ctx.Errors[0]

		if errors.Is(err, &BadRequestError{}) || errors.Is(err, validator.ValidationErrors{}) {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			if ctx.Request.Method == http.MethodDelete {
				ctx.Status(http.StatusNoContent)
				return
			}
			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, os.ErrNotExist) {
			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				})
			return
		}

		if errors.Is(err, model.DependencyCyclicError{}) {
			ctx.JSON(
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
				ctx.JSON(
					http.StatusConflict,
					gin.H{
						"error": err.Error(),
					})
				return
			}
		}

		ctx.JSON(
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
