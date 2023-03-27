package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	liberr "github.com/konveyor/controller/pkg/error"
	"reflect"
	"runtime"
)

//
// HandlerPositionError report handler found in invalid position.
type HandlerPositionError struct {
	Name     string
	Position int
	Chain    []string
}

func (r *HandlerPositionError) Error() string {
	return fmt.Sprintf(
		"Invalid handler chain: %s."+
			"%s found at: %d/%d",
		r.Chain,
		r.Name,
		r.Position,
		len(r.Chain))
}

//
// WatchDog handler inspect the handler chain.
// Inspections:
//  - Transaction may only appear next to last.
func WatchDog() gin.HandlerFunc {
	fv := reflect.ValueOf(Transaction)
	fp := fv.Pointer()
	txHandler := runtime.FuncForPC(fp).Name()
	return func(ctx *gin.Context) {
		chain := ctx.HandlerNames()
		for i := range chain {
			if i != len(chain)-2 && chain[i] == txHandler {
				err := &HandlerPositionError{
					Name:     txHandler,
					Position: i,
					Chain:    chain,
				}
				_ = ctx.Error(liberr.Wrap(err))
				ctx.Abort()
				return
			}
		}
	}
}
