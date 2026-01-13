package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/shared/api"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelConfigMap = "konveyor.io/configuration"
)

// ConfigMapHandler handles configmap routes.
type ConfigMapHandler struct {
	BaseHandler
}

// AddRoutes add routes.
func (h ConfigMapHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("configmaps"))
	routeGroup.GET(api.ConfigMapsRoute, h.List)
	routeGroup.GET(api.ConfigMapsRoute+"/", h.List)
	routeGroup.GET(api.ConfigMapRoute, h.Get)
	routeGroup.GET(api.ConfigMapKeyRoute, h.Get)
}

// Get godoc
// @summary Get a configmap by name.
// @description Get a configmap by name.
// @tags ConfigMaps
// @produce json
// @success 200 {object} api.ConfigMap
// @router /configmaps/{name} [get]
// @param name path string true "Name"
func (h ConfigMapHandler) Get(ctx *gin.Context) {
	name := ctx.Param(Name)
	key := ctx.Param(Key)
	mp := &v1.ConfigMap{}
	err := h.Client(ctx).Get(
		context.TODO(),
		k8s.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		mp)
	if err != nil {
		if errors.IsNotFound(err) {
			h.Status(ctx, http.StatusNotFound)
			return
		} else {
			_ = ctx.Error(err)
			return
		}
	}
	if _, found := mp.Labels[LabelConfigMap]; !found {
		h.Status(ctx, http.StatusNotFound)
		return
	}
	var r any
	var found bool
	r = ConfigMap{
		Name: mp.Name,
		Data: mp.Data,
	}
	if key != "" {
		r, found = mp.Data[key]
		if !found {
			h.Status(ctx, http.StatusNotFound)
			return
		}
	}

	h.Respond(ctx, http.StatusOK, r)
}

// GetKey godoc
// @summary Get a configmap by name and key.
// @description Get a configmap by name and key.
// @tags ConfigMaps
// @produce json
// @success 200 {string} string
// @router /configmaps/{name} [get]
// @param name path string true "Name"
// @param key path string true "Key"
func (h ConfigMapHandler) GetKey(ctx *gin.Context) {
	h.Get(ctx)
}

// List godoc
// @summary List all configmap names.
// @description List all configmap names.
// @tags ConfigMaps
// @produce json
// @success 200 {array} api.ConfigMap
// @router /configmaps [get]
func (h ConfigMapHandler) List(ctx *gin.Context) {
	maps := &v1.ConfigMapList{}
	selector := labels.NewSelector()
	req, _ := labels.NewRequirement(
		LabelConfigMap,
		selection.Exists,
		[]string{})
	selector = selector.Add(*req)
	err := h.Client(ctx).List(
		context.TODO(),
		maps,
		&k8s.ListOptions{
			Namespace:     Settings.Namespace,
			LabelSelector: selector,
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var resources []ConfigMap
	for _, m := range maps.Items {
		resources = append(
			resources,
			ConfigMap{
				Name: m.Name,
				Data: m.Data,
			})
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// ConfigMap configmap
type ConfigMap = resource.ConfigMap
