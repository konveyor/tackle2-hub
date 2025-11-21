package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// Routes
const (
	ConfigurationsRoot   = "/configurations"
	ConfigurationRoot    = ConfigurationsRoot + "/:" + Name
	ConfigurationKeyRoot = ConfigurationRoot + "/:" + Key
)

const (
	LabelConfiguration = "konveyor.io/configuration"
)

// ConfigurationHandler handles configuration routes.
type ConfigurationHandler struct {
	BaseHandler
}

// AddRoutes add routes.
func (h ConfigurationHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("Configurations"))
	routeGroup.GET(ConfigurationsRoot, h.List)
	routeGroup.GET(ConfigurationsRoot+"/", h.List)
	routeGroup.GET(ConfigurationRoot, h.Get)
	routeGroup.GET(ConfigurationKeyRoot, h.Get)
}

// Get godoc
// @summary Get a configuration by name.
// @description Get a configuration by name.
// @tags Configurations
// @produce json
// @success 200 {object} any
// @router /Configurations/{name} [get]
// @param name path string true "Name"
func (h ConfigurationHandler) Get(ctx *gin.Context) {
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
	if _, found := mp.Labels[LabelConfiguration]; !found {
		h.Status(ctx, http.StatusNotFound)
		return
	}
	var r any
	var found bool
	r = mp.Data
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
// @summary Get a configuration by name and key.
// @description Get a configuration by name and key.
// @tags Configurations
// @produce json
// @success 200 {object} any
// @router /Configurations/{name} [get]
// @param name path string true "Key"
// @param key path string true "Name"
func (h ConfigurationHandler) GetKey(ctx *gin.Context) {
	h.Get(ctx)
}

// List godoc
// @summary List all configuration names.
// @description List all configuration names.
// @tags Configurations
// @produce json
// @success 200 array api.Configuration
// @router /Configurations [get]
func (h ConfigurationHandler) List(ctx *gin.Context) {
	maps := &v1.ConfigMapList{}
	selector := labels.NewSelector()
	req, _ := labels.NewRequirement(
		LabelConfiguration,
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
	var resources []Configuration
	for _, m := range maps.Items {
		resources = append(
			resources,
			Configuration{
				Name: m.Name,
				Data: m.Data,
			})
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Configuration configuration
type Configuration struct {
	Name string `json:"names"`
	Data any    `json:"data"`
}
