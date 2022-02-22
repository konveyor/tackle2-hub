package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
	"strconv"
)

//
// Kind
const (
	ApplicationKind = "application"
)

//
// Routes
const (
	ApplicationsRoot = InventoryRoot + "/application"
	ApplicationRoot  = ApplicationsRoot + "/:" + ID
)

//
// ApplicationHandler handles application resource routes.
type ApplicationHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h ApplicationHandler) AddRoutes(e *gin.Engine) {
	e.GET(ApplicationsRoot, h.List)
	e.GET(ApplicationsRoot+"/", h.List)
	e.POST(ApplicationsRoot, h.Create)
	e.GET(ApplicationRoot, h.Get)
	e.PUT(ApplicationRoot, h.Update)
	e.DELETE(ApplicationRoot, h.Delete)
}

// Get godoc
// @summary Get an application by ID.
// @description Get an application by ID.
// @tags get
// @produce json
// @success 200 {object} api.Application
// @router /application-inventory/application/{id} [get]
// @param id path int true "Application ID"
func (h ApplicationHandler) Get(ctx *gin.Context) {
	m := &model.Application{}
	id := ctx.Param(ID)
	db := h.preLoad(
		h.DB,
		"Tags",
		"Review",
		"BusinessService")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Application{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all applications.
// @description List all applications.
// @tags list
// @produce json
// @success 200 {object} []api.Application
// @router /application-inventory/application [get]
func (h ApplicationHandler) List(ctx *gin.Context) {
	var count int64
	var list []model.Application
	h.DB.Model(model.Application{}).Count(&count)
	pagination := NewPagination(ctx)
	db := pagination.apply(h.DB)
	db = h.BaseHandler.preLoad(
		db,
		"Tags",
		"Review",
		"BusinessService")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Application{}
	for i := range list {
		r := Application{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.listResponse(ctx, ApplicationKind, resources, int(count))
}

// Create godoc
// @summary Create an application.
// @description Create an application.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Application
// @router /application-inventory/application [post]
// @param application body api.Application true "Application data"
func (h ApplicationHandler) Create(ctx *gin.Context) {
	r := &Application{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	err = h.DB.Model(m).Association("Tags").Replace("Tags", m.Tags)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an application.
// @description Delete an application.
// @tags delete
// @success 204
// @router /application-inventory/application/{id} [delete]
// @param id path int true "Application id"
func (h ApplicationHandler) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param(ID))
	m := &model.Application{}
	m.ID = uint(id)
	result := h.DB.Select("Tags").Delete(m)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update an application.
// @description Update an application.
// @tags update
// @accept json
// @success 204
// @router /application-inventory/application/{id} [put]
// @param id path int true "Application id"
// @param application body api.Application true "Application data"
func (h ApplicationHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	r := &Application{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Model(&model.Application{}).Where("id = ?", id).Omit("id").Updates(m)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	err = h.DB.Model(&m).Association("Tags").Replace("Tags", m.Tags)
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Application REST resource.
type Application struct {
	Resource
	Name            string      `json:"name" binding:"required"`
	Description     string      `json:"description"`
	Repository      *Repository `json:"repository"`
	Extensions      Extensions  `json:"extensions"`
	Review          *Review     `json:"review"`
	Comments        string      `json:"comments"`
	Tags            []string    `json:"tags"`
	BusinessService string      `json:"businessService"`
}

//
// With updates the resource using the model.
func (r *Application) With(m *model.Application) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Comments = m.Comments
	_ = json.Unmarshal(m.Repository, &r.Repository)
	_ = json.Unmarshal(m.Extensions, &r.Extensions)
	if m.Review != nil {
		r.Review = &Review{Resource: Resource{ID: m.Review.ID}}
	}
	r.BusinessService = strconv.Itoa(int(m.BusinessServiceID))
	for _, tag := range m.Tags {
		r.Tags = append(
			r.Tags,
			strconv.Itoa(int(tag.ID)))
	}
}

//
// Model builds a model.
func (r *Application) Model() (m *model.Application) {
	m = &model.Application{
		Name:        r.Name,
		Description: r.Description,
		Comments:    r.Comments,
	}
	m.ID = r.ID
	if r.Repository != nil {
		m.Repository, _ = json.Marshal(r.Repository)
	}
	if r.Extensions != nil {
		m.Extensions, _ = json.Marshal(r.Extensions)
	}
	if len(r.BusinessService) > 0 {
		id, _ := strconv.Atoi(r.BusinessService)
		m.BusinessServiceID = uint(id)
	}
	for _, tagID := range r.Tags {
		id, _ := strconv.Atoi(tagID)
		m.Tags = append(
			m.Tags,
			model.Tag{
				Model: model.Model{
					ID: uint(id),
				},
			})
	}

	return
}

//
// Repository REST nested resource.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url" binding:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

//
// Extensions of the application.
type Extensions map[string]interface{}
