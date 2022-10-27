package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

//
// Routes
const (
	AddonsRoot = "/addons"
	AddonRoot  = AddonsRoot + "/:" + Name
)

//
// AddonHandler handles addon routes.
type AddonHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h AddonHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("addons"))
	routeGroup.GET(AddonsRoot, h.List)
	routeGroup.GET(AddonsRoot+"/", h.List)
	routeGroup.GET(AddonRoot, h.Get)
}

// Get godoc
// @summary Get an addon by name.
// @description Get an addon by name.
// @tags get
// @produce json
// @success 200 {object} api.Addon
// @router /addons/{name} [get]
// @param name path string true "Addon name"
func (h AddonHandler) Get(ctx *gin.Context) {
	name := ctx.Param(Name)
	addon := &crd.Addon{}
	err := h.Client.Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		addon)
	if err != nil {
		if errors.IsNotFound(err) {
			ctx.Status(http.StatusNotFound)
			return
		} else {
			h.getFailed(ctx, err)
			return
		}
	}
	r := Addon{}
	r.With(addon)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all addons.
// @description List all addons.
// @tags get
// @produce json
// @success 200 {object} []api.Addon
// @router /addons [get]
func (h AddonHandler) List(ctx *gin.Context) {
	list := &crd.AddonList{}
	err := h.Client.List(
		context.TODO(),
		&k8s.ListOptions{
			Namespace: Settings.Namespace,
		},
		list)
	if err != nil {
		h.listFailed(ctx, err)
		return
	}
	content := []Addon{}
	for _, m := range list.Items {
		addon := Addon{}
		addon.With(&m)
		content = append(content, addon)
	}

	ctx.JSON(http.StatusOK, content)
}

//
// Addon REST resource.
type Addon struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

//
// With model.
func (r *Addon) With(m *crd.Addon) {
	r.Name = m.Name
	r.Image = m.Spec.Image
}
