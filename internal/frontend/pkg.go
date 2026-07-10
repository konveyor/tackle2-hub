package frontend

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/frontend/auth"
)

type Handler interface {
	AddRoutes(*gin.Engine)
}

func ALL() []Handler {
	return []Handler{
		auth.Handler{},
	}
}
