package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

const (
	Kind = "kind"
)

// ProxyHandler handles proxy resource routes.
type ProxyHandler struct {
	BaseHandler
}

func (h ProxyHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("proxies"))
	routeGroup.GET(api.ProxiesRoute, h.List)
	routeGroup.GET(api.ProxiesRoute+"/", h.List)
	routeGroup.POST(api.ProxiesRoute, h.Create)
	routeGroup.GET(api.ProxyRoute, h.Get)
	routeGroup.PUT(api.ProxyRoute, h.Update)
	routeGroup.DELETE(api.ProxyRoute, h.Delete)
}

// Get godoc
// @summary Get an proxy by ID.
// @description Get an proxy by ID.
// @tags proxies
// @produce json
// @success 200 {object} Proxy
// @router /proxies/{id} [get]
// @param id path int true "Proxy ID"
func (h ProxyHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	proxy := &model.Proxy{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(proxy, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Proxy{}
	r.With(proxy)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all proxies.
// @description List all proxies.
// @tags proxies
// @produce json
// @success 200 {object} []Proxy
// @router /proxies [get]
func (h ProxyHandler) List(ctx *gin.Context) {
	var list []model.Proxy
	kind := ctx.Query(Kind)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	if kind != "" {
		db = db.Where(Kind, kind)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Proxy{}
	for i := range list {
		r := Proxy{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create an proxy.
// @description Create an proxy.
// @tags proxies
// @accept json
// @produce json
// @success 201 {object} Proxy
// @router /proxies [post]
// @param proxy body Proxy true "Proxy data"
func (h ProxyHandler) Create(ctx *gin.Context) {
	proxy := &Proxy{}
	err := h.Bind(ctx, proxy)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := proxy.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	proxy.With(m)

	h.Respond(ctx, http.StatusCreated, proxy)
}

// Delete godoc
// @summary Delete an proxy.
// @description Delete an proxy.
// @tags proxies
// @success 204
// @router /proxies/{id} [delete]
// @param id path int true "Proxy ID"
func (h ProxyHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	proxy := &model.Proxy{}
	result := h.DB(ctx).First(proxy, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(proxy, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update an proxy.
// @description Update an proxy.
// @tags proxies
// @accept json
// @success 204
// @router /proxies/{id} [put]
// @param id path int true "Proxy ID"
// @param proxy body Proxy true "Proxy data"
func (h ProxyHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Proxy{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result := db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Proxy REST resource.
type Proxy = resource.Proxy
