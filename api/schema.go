package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/jsd"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
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
// @param name path int true "Schema name"
func (h *SchemaHandler) Get(ctx *gin.Context) {
	name := ctx.Param(Name)
	m := &crd.Schema{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		m)
	if err != nil {
		if errors.IsNotFound(err) {
			h.Status(ctx, http.StatusNotFound)
			return
		} else {
			_ = ctx.Error(err)
			return
		}
	}
	r := Schema{}
	r.With(m)
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
	list := &crd.SchemaList{}
	err := h.Client(ctx).List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	content := []Schema{}
	for _, m := range list.Items {
		r := Schema{}
		r.With(&m)
		content = append(content, r)
	}

	h.Respond(ctx, http.StatusOK, content)
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

type Schema struct {
	Name     string           `json:"name"`
	Domain   string           `json:"domain"`
	Variant  string           `json:"variant"`
	Subject  string           `json:"subject"`
	Versions jsd.Versions     `json:"versions"`
	Status   crd.SchemaStatus `json:"status,omitempty"`
}

func (r *Schema) With(m *crd.Schema) {
	r.Name = m.Name
	r.Domain = m.Spec.Domain
	r.Variant = m.Spec.Variant
	r.Subject = m.Spec.Subject
	r.Versions = make(jsd.Versions, 0)
	for id := range m.Spec.Versions {
		v := m.Spec.Versions[id]
		definition := make(Map)
		_ = json.Unmarshal(v.Definition.Raw, &definition)
		r.Versions = append(
			r.Versions,
			jsd.Version{
				ID:         id,
				Migration:  v.Migration,
				Definition: definition,
			})
	}
}

type LatestSchema struct {
	Name       string `json:"name"`
	Definition Map    `json:"definition"`
}
