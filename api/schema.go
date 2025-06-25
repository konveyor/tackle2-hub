package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/jsd"
)

const (
	Domain  = "domain"
	Variant = "variant"
	Subject = "subject"
)

const (
	SchemaRoot  = "/schema"
	SchemasRoot = "/schemas" + "/:" + Domain + "/:" + Variant + "/:" + Subject
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
	r.GET(SchemaRoot, h.GetAPI)
}

// GetAPI godoc
// @summary Get the API schema.
// @description Get the API schema.
// @tags schema
// @produce json
// @success 200 {object} RestAPI
// @router /schema [get]
func (h *SchemaHandler) GetAPI(ctx *gin.Context) {
	api := RestAPI{
		Version: h.Version,
		Paths:   []string{},
	}
	for _, rte := range h.router.Routes() {
		api.Paths = append(api.Paths, rte.Path)
	}

	h.Respond(ctx, http.StatusOK, api)
}

// Get godoc
// @summary Get a schema.
// @description Get a schema.
// @tags schema
// @produce json
// @success 200 {object} Schema
// @router /schema [get]
func (h *SchemaHandler) Get(ctx *gin.Context) {
	domain := ctx.Param(Domain)
	variant := ctx.Param(Variant)
	subject := ctx.Param(Subject)
	m := jsd.Manager{Client: h.Client(ctx)}
	v, err := m.Latest(domain, variant, subject)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Schema{
		Domain:  domain,
		Variant: variant,
		Subject: subject,
		Version: v.Id,
		Content: v.Content,
	}
	h.Respond(ctx, http.StatusOK, r)
}

type RestAPI struct {
	Version string   `json:"version,omitempty"`
	Paths   []string `json:"paths"`
}

type Schema struct {
	Domain  string `json:"domain"`
	Variant string `json:"variant"`
	Subject string `json:"subject"`
	Version int    `json:"version,omitempty"`
	Content Map    `json:"content,omitempty"`
}
