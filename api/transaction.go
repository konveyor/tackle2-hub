package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//
// Context keys.
const (
	CtxDB     = "DB"
	CtxClient = "CLIENT"
)

//
// TxHandler DB transaction handler.
func TxHandler(db *gorm.DB, c client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxClient, c)
		switch ctx.Request.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete:
			err := db.Transaction(func(tx *gorm.DB) (err error) {
				ctx.Set(CtxDB, tx)
				ctx.Next()
				if len(ctx.Errors) > 0 {
					err = ctx.Errors[0]
				}
				return
			})
			if err != nil {
				_ = ctx.Error(err)
			}
		default:
			ctx.Set(CtxDB, db)
			ctx.Next()
		}
	}
}
