package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
)

// Routes
const (
	PlatformsRoot = "/platforms"
	PlatformRoot  = PlatformsRoot + "/:" + ID
)

// PlatformHandler handles application Platform resource routes.
type PlatformHandler struct {
	BaseHandler
}

func (h PlatformHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("platforms"))
	routeGroup.GET(PlatformRoot, h.Get)
	routeGroup.GET(PlatformsRoot, h.List)
	routeGroup.GET(PlatformsRoot+"/", h.List)
	routeGroup.POST(PlatformsRoot, h.Create)
	routeGroup.PUT(PlatformRoot, h.Update)
	routeGroup.DELETE(PlatformRoot, h.Delete)
}

// Get godoc
// @summary Get a Platform by ID.
// @description Get a Platform by ID.
// @tags platforms
// @produce json
// @success 200 {object} Platform
// @router /platforms/{id} [get]
// @param id path int true "Platform ID"
func (h PlatformHandler) Get(ctx *gin.Context) {
	r := Platform{}
	id := h.pk(ctx)
	m := &model.Platform{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all platforms.
// @description List all platforms.
// @tags platforms
// @produce json
// @success 200 {object} []Platform
// @router /platforms [get]
func (h PlatformHandler) List(ctx *gin.Context) {
	resources := []Platform{}
	var list []model.Platform
	db := h.DB(ctx)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		r := Platform{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a platform.
// @description Create a platform.
// @tags platforms
// @accept json
// @produce json
// @success 201 {object} Platform
// @router /platforms [post]
// @param platform body Platform true "Platform data"
func (h PlatformHandler) Create(ctx *gin.Context) {
	r := &Platform{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a platform.
// @description Delete a platform.
// @tags platforms
// @success 204
// @router /platforms/{id} [delete]
// @param id path int true "Platform ID"
func (h PlatformHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Platform{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = db.Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a platform.
// @description Update a platform.
// @tags platforms
// @accept json
// @success 204
// @router /platforms/{id} [put]
// @param id path int true "Platform ID"
// @param platform body Platform true "Platform data"
func (h PlatformHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Platform{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Platform REST resource.
type Platform struct {
	Resource     `yaml:",inline"`
	Kind         string `json:"kind" binding:"required"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Identity     *Ref   `json:"identity,omitempty" yaml:",omitempty"`
	Applications []Ref  `json:"applications,omitempty" yaml:",omitempty"`
}

// With updates the resource with the model.
func (r *Platform) With(m *model.Platform) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.URL = m.URL
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	r.Applications = make([]Ref, len(m.Applications))
	for _, application := range m.Applications {
		r.Applications = append(
			r.Applications,
			r.ref(application.ID, application.ID))
	}
}

// Model builds a model.
func (r *Platform) Model() (m *model.Platform) {
	m = &model.Platform{}
	m.ID = r.ID
	m.Kind = r.Kind
	m.Name = r.Name
	m.URL = r.URL
	if r.Identity != nil {
		m.IdentityID = &r.Identity.ID
	}
	return
}
