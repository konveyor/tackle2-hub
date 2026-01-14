package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/jsd"
	"github.com/konveyor/tackle2-hub/shared/api"
)

const (
	Domain  = api.Domain
	Variant = api.Variant
	Subject = api.Subject
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
	r.GET(api.SchemaRoute, h.GetAPI)
	r.GET(api.SchemasRoute, h.List)
	r.GET(api.SchemasGetRoute, h.Get)
	r.GET(api.SchemaFindRoute, h.Find)
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
	r := Schema{}
	r.With(s)
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
		s := Schema{}
		s.With(list[i])
		r[i] = s
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
		Definition: resource.Map(v.Definition),
		Name:       s.Name,
	}
	h.Respond(ctx, http.StatusOK, r)
}

// RestAPI resource.
type RestAPI = resource.RestAPI

// Schema REST resource.
type Schema = resource.Schema

// LatestSchema REST resource.
type LatestSchema = resource.LatestSchema
