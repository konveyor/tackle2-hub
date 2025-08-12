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
	SchemaRoot     = "/schema"
	SchemasRoot    = "/schemas"
	SchemasGetRoot = SchemasRoot + "/:" + Name
	SchemaFindRoot = SchemaRoot + "/jsd/:" + Domain + "/:" + Variant + "/:" + Subject
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
	r.GET(SchemaRoot, h.GetAPI)
	r.GET(SchemasRoot, h.List)
	r.GET(SchemasGetRoot, h.Get)
	r.GET(SchemaFindRoot, h.Find)
}

// GetAPI godoc
// @summary Get the API routes.
// @description Get the API routes.
// @tags schema
// @produce json
// @success 200 {object} RestAPI
// @router /schema [get]
func (h *SchemaHandler) GetAPI(ctx *gin.Context) {
	api := RestAPI{
		Version: h.Version,
		Routes:  []string{},
	}
	for _, rte := range h.router.Routes() {
		api.Routes = append(api.Routes, rte.Path)
	}

	h.Respond(ctx, http.StatusOK, api)
}

// Get godoc
// @summary Find a schema.
// @description Find a schema.
// @tags schemas
// @produce json
// @success 200 {object} Schema
// @router /schemas/{name} [get]
// @param name path string true "Schema name"
func (h *SchemaHandler) Get(ctx *gin.Context) {
	name := ctx.Param(Name)
	m := jsd.Manager{Client: h.Client(ctx)}
	s, err := m.Get(name)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Schema(s)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List schemas.
// @description List schemas.
// @tags schema
// @produce json
// @success 200 {object} []Schema
// @router /schemas [get]
func (h *SchemaHandler) List(ctx *gin.Context) {
	m := jsd.Manager{Client: h.Client(ctx)}
	list, err := m.List()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := make([]Schema, len(list))
	for i := range list {
		r[i] = Schema(list[i])
	}
	h.Respond(ctx, http.StatusOK, r)
}

// Find godoc
// @summary Find a schema.
// @description Find a schema.
// @tags schema
// @produce json
// @success 200 {object} LatestSchema
// @router /schema/jsd/{domain}/{variant}/{subject} [get]
// @param domain path string true "The schema domain."
// @param variant path string true "The schema variant."
// @param subject path string true "The schema subject."
func (h *SchemaHandler) Find(ctx *gin.Context) {
	domain := ctx.Param(Domain)
	variant := ctx.Param(Variant)
	subject := ctx.Param(Subject)
	m := jsd.Manager{Client: h.Client(ctx)}
	s, err := m.Find(domain, variant, subject)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if len(s.Versions) == 0 {
		_ = ctx.Error(&jsd.NotFound{})
		return
	}
	v := s.Versions.Latest()
	r := LatestSchema{
		Definition: Map(v.Definition),
		Name:       s.Name,
	}
	h.Respond(ctx, http.StatusOK, r)
}

type RestAPI struct {
	Version string   `json:"version,omitempty" yaml:",omitempty"`
	Routes  []string `json:"routes"`
}

type Schema jsd.Schema

type LatestSchema struct {
	Name       string `json:"name"`
	Definition Map    `json:"definition"`
}
