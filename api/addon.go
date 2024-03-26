package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// Routes
const (
	AddonsRoot = "/addons"
	AddonRoot  = AddonsRoot + "/:" + Name
)

// AddonHandler handles addon routes.
type AddonHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AddonHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("addons"))
	routeGroup.GET(AddonsRoot, h.List)
	routeGroup.GET(AddonsRoot+"/", h.List)
	routeGroup.GET(AddonRoot, h.Get)
}

// Get godoc
// @summary Get an addon by name.
// @description Get an addon by name.
// @tags addons
// @produce json
// @success 200 {object} api.Addon
// @router /addons/{name} [get]
// @param name path string true "Addon name"
func (h AddonHandler) Get(ctx *gin.Context) {
	name := ctx.Param(Name)
	addon := &crd.Addon{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		addon)
	if err != nil {
		if errors.IsNotFound(err) {
			h.Status(ctx, http.StatusNotFound)
			return
		} else {
			_ = ctx.Error(err)
			return
		}
	}
	extensions := &crd.ExtensionList{}
	err = h.Client(ctx).List(
		context.TODO(),
		extensions,
		&k8s.ListOptions{
			Namespace: Settings.Namespace,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Addon{}
	r.With(addon, extensions.Items...)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all addons.
// @description List all addons.
// @tags addons
// @produce json
// @success 200 {object} []api.Addon
// @router /addons [get]
func (h AddonHandler) List(ctx *gin.Context) {
	list := &crd.AddonList{}
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
	content := []Addon{}
	for _, m := range list.Items {
		addon := Addon{}
		addon.With(&m)
		content = append(content, addon)
	}

	h.Respond(ctx, http.StatusOK, content)
}

// Addon REST resource.
type Addon struct {
	Name       string         `json:"name"`
	Capability string         `json:"capability,omitempty"`
	Container  core.Container `json:"container"`
	Extensions []Extension    `json:"extensions,omitempty"`
	Metadata   any            `json:"metadata,omitempty"`
}

// With model.
func (r *Addon) With(m *crd.Addon, extensions ...crd.Extension) {
	r.Name = m.Name
	r.Capability = m.Spec.Capability
	r.Container = m.Spec.Container
	if m.Spec.Metadata.Raw != nil {
		_ = json.Unmarshal(m.Spec.Metadata.Raw, &r.Metadata)
	}
	for i := range extensions {
		extension := Extension{}
		extension.With(&extensions[i])
		r.Extensions = append(
			r.Extensions,
			extension)
	}
}

// Extension REST resource.
type Extension struct {
	Name       string         `json:"name"`
	Addon      string         `json:"addon"`
	Capability string         `json:"capability,omitempty"`
	Container  core.Container `json:"container"`
	Metadata   any            `json:"metadata,omitempty"`
}

// With model.
func (r *Extension) With(m *crd.Extension) {
	r.Name = m.Name
	r.Addon = m.Spec.Capability
	r.Capability = m.Spec.Capability
	r.Container = m.Spec.Container
	if m.Spec.Metadata.Raw != nil {
		_ = json.Unmarshal(m.Spec.Metadata.Raw, &r.Metadata)
	}
}
