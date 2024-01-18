package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// SchemaHandler providers schema (route) handler.
type SchemaHandler struct {
	BaseHandler
	// The `gin` router.
	router *gin.Engine
	// Schema version
	Version string
}

// AddRoutes Adds routes.
func (h *SchemaHandler) AddRoutes(r *gin.Engine) {
	h.router = r
	//
	// Routes
	r.GET("/schema", h.Get)
}

// Get godoc
// @summary Get the API schema.
// @description Get the API schema.
// @tags schema
// @produce json
// @success 200 {object} Schema
// @router /schema [get]
func (h *SchemaHandler) Get(ctx *gin.Context) {
	schema := Schema{
		Version: h.Version,
		Paths:   []string{},
	}
	for _, rte := range h.router.Routes() {
		schema.Paths = append(schema.Paths, rte.Path)
	}

	h.Respond(ctx, http.StatusOK, schema)
}

type Schema struct {
	Version string   `json:"version,omitempty"`
	Paths   []string `json:"paths"`
}
