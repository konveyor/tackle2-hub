package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Routes
const (
	GeneratorsRoot = "/generators"
	GeneratorRoot  = GeneratorsRoot + "/:" + ID
)

// GeneratorHandler handles application Generator resource routes.
type GeneratorHandler struct {
	BaseHandler
}

func (h GeneratorHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("generators"))
	routeGroup.GET(GeneratorRoot, h.Get)
	routeGroup.GET(GeneratorsRoot, h.List)
	routeGroup.GET(GeneratorsRoot+"/", h.List)
	routeGroup.POST(GeneratorsRoot, h.Create)
	routeGroup.PUT(GeneratorRoot, h.Update)
	routeGroup.DELETE(GeneratorRoot, h.Delete)
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
type Generator struct {
	Resource   `yaml:",inline"`
	Kind       string      `json:"kind" binding:"required"`
	Name       string      `json:"name"`
	Repository *Repository `json:"repository"`
	Parameters Map         `json:"parameters"`
	Values     Map         `json:"values"`
	Identity   *Ref        `json:"identity,omitempty" yaml:",omitempty"`
	Profiles   []Ref       `json:"profiles"`
}

// With updates the resource with the model.
func (r *Generator) With(m *model.Generator) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	r.Parameters = m.Parameters
	r.Values = m.Values
	if m.Repository != (model.Repository{}) {
		repository := Repository(m.Repository)
		r.Repository = &repository
	}
	r.Profiles = make([]Ref, 0, len(m.Profiles))
	for _, p := range m.Profiles {
		r.Profiles = append(
			r.Profiles,
			r.ref(p.ID, &p))
	}
}

// Model builds a model.
func (r *Generator) Model() (m *model.Generator) {
	m = &model.Generator{}
	m.ID = r.ID
	m.Kind = r.Kind
	m.Name = r.Name
	m.Parameters = r.Parameters
	m.Values = r.Values
	if r.Repository != nil {
		m.Repository = model.Repository(*r.Repository)
	}
	if r.Identity != nil {
		m.IdentityID = &r.Identity.ID
	}
	return
}
