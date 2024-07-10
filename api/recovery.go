package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/lifecycle"
)

const (
	StopRoute  = "/service/stop"
	StartRoute = "/service/start"
)

type RecoveryHandler struct {
	Manager *lifecycle.Manager
}

func (h RecoveryHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.POST(StopRoute, h.Stop)
	routeGroup.POST(StartRoute, h.Start)
}

func (h RecoveryHandler) Stop(ctx *gin.Context) {
	h.Manager.Stop()
	ctx.Status(202)
}

func (h RecoveryHandler) Start(ctx *gin.Context) {
	h.Manager.Run()
	ctx.Status(202)
}
