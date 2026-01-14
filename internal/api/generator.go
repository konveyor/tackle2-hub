package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// GeneratorHandler handles application Generator resource routes.
type GeneratorHandler struct {
	BaseHandler
}

func (h GeneratorHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("generators"))
	routeGroup.GET(api.GeneratorRoute, h.Get)
	routeGroup.GET(api.GeneratorsRoute, h.List)
	routeGroup.GET(api.GeneratorsRoute+"/", h.List)
	routeGroup.POST(api.GeneratorsRoute, h.Create)
	routeGroup.PUT(api.GeneratorRoute, h.Update)
	routeGroup.DELETE(api.GeneratorRoute, h.Delete)
}

// Get godoc
// @summary Get a Generator by ID.
// @description Get a Generator by ID.
// @tags generators
// @produce json
// @success 200 {object} Generator
// @router /generators/{id} [get]
// @param id path int true "Generator ID"
func (h GeneratorHandler) Get(ctx *gin.Context) {
	r := Generator{}
	id := h.pk(ctx)
	m := &model.Generator{}
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all generators.
// @description List all generators.
// @tags generators
// @produce json
// @success 200 {object} []Generator
// @router /generators [get]
func (h GeneratorHandler) List(ctx *gin.Context) {
	resources := []Generator{}
	var list []model.Generator
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		r := Generator{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a generator.
// @description Create a generator.
// @tags generators
// @accept json
// @produce json
// @success 201 {object} Generator
// @router /generators [post]
// @param generator body Generator true "Generator data"
func (h GeneratorHandler) Create(ctx *gin.Context) {
	r := &Generator{}
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
// @summary Delete a generator.
// @description Delete a generator.
// @tags generators
// @success 204
// @router /generators/{id} [delete]
// @param id path int true "Generator ID"
func (h GeneratorHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Generator{}
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
// @summary Update a generator.
// @description Update a generator.
// @tags generators
// @accept json
// @success 204
// @router /generators/{id} [put]
// @param id path int true "Generator ID"
// @param generator body Generator true "Generator data"
func (h GeneratorHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Generator{}
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

// Generator REST resource.
type Generator = resource.Generator
