package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	ApplicationsRoot = "/applications"
	ApplicationRoot  = ApplicationsRoot + "/:" + ID
	AppBucketRoot    = ApplicationRoot + "/bucket/*" + Wildcard
)

//
// ApplicationHandler handles application resource routes.
type ApplicationHandler struct {
	BaseHandler
	BucketHandler
}

//
// AddRoutes adds routes.
func (h ApplicationHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, "applications"))
	routeGroup.GET(ApplicationsRoot, h.List)
	routeGroup.GET(ApplicationsRoot+"/", h.List)
	routeGroup.POST(ApplicationsRoot, h.Create)
	routeGroup.GET(ApplicationRoot, h.Get)
	routeGroup.PUT(ApplicationRoot, h.Update)
	routeGroup.DELETE(ApplicationsRoot, h.DeleteList)
	routeGroup.DELETE(ApplicationRoot, h.Delete)
	routeGroup.POST(AppBucketRoot, h.BucketUpload)
	routeGroup.PUT(AppBucketRoot, h.BucketUpload)
	routeGroup.GET(AppBucketRoot, h.BucketGet)
	routeGroup.DELETE(AppBucketRoot, h.BucketDelete)
}

// Get godoc
// @summary Get an application by ID.
// @description Get an application by ID.
// @tags get
// @produce json
// @success 200 {object} api.Application
// @router /applications/{id} [get]
// @param id path int true "Application ID"
func (h ApplicationHandler) Get(ctx *gin.Context) {
	m := &model.Application{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB, clause.Associations)
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
// @router /applications [get]
func (h ApplicationHandler) List(ctx *gin.Context) {
	var list []model.Application
	db := h.preLoad(h.DB, clause.Associations)
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

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create an application.
// @description Create an application.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Application
// @router /applications [post]
// @param application body api.Application true "Application data"
func (h ApplicationHandler) Create(ctx *gin.Context) {
	r := &Application{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	err = h.DB.Model(m).Association("Identities").Replace("Identities", m.Identities)
	if err != nil {
		h.createFailed(ctx, err)
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
// @router /applications/{id} [delete]
// @param id path int true "Application id"
func (h ApplicationHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Application{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}
	result = h.DB.Delete(m)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// DeleteList godoc
// @summary Delete a applications.
// @description Delete applications.
// @tags delete
// @success 204
// @router /applications [delete]
// @param application body []uint true "List of id"
func (h ApplicationHandler) DeleteList(ctx *gin.Context) {
	ids := []uint{}
	err := ctx.BindJSON(&ids)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	err = h.DB.Delete(
		&model.Application{},
		"id IN ?",
		ids).Error
	if err != nil {
		h.deleteFailed(ctx, err)
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
// @router /applications/{id} [put]
// @param id path int true "Application id"
// @param application body api.Application true "Application data"
func (h ApplicationHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Application{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB.Model(m)
	db = db.Omit(clause.Associations)
	db = db.Omit("Bucket")
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	db = h.DB.Model(m)
	err = db.Association("Identities").Replace(m.Identities)
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}
	db = h.DB.Model(m)
	err = db.Association("Tags").Replace(m.Tags)
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// BucketGet godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @tags get
// @produce octet-stream
// @success 200
// @router /applications/{id}/tasks/{id}/content/{wildcard} [get]
// @param id path string true "Task ID"
func (h ApplicationHandler) BucketGet(ctx *gin.Context) {
	taskID := h.pk(ctx)
	m := &model.Application{}
	result := h.DB.First(m, taskID)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	h.content(ctx, &m.BucketOwner)
}

// BucketUpload godoc
// @summary Upload bucket content by ID and path.
// @description Upload bucket content by ID and path.
// @tags post
// @produce json
// @success 204
// @router /applications/{id}/bucket/{wildcard} [post]
// @param id path string true "Application ID"
func (h ApplicationHandler) BucketUpload(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Application{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	h.upload(ctx, &m.BucketOwner)
}

// BucketDelete godoc
// @summary Delete bucket content by ID and path.
// @description Delete bucket content by ID and path.
// @tags delete
// @produce json
// @success 204
// @router /applications/{id}/bucket/{wildcard} [delete]
// @param id path string true "Application ID"
func (h ApplicationHandler) BucketDelete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Application{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	h.delete(ctx, &m.BucketOwner)
}

//
// Application REST resource.
type Application struct {
	Resource
	Name            string      `json:"name" binding:"required"`
	Description     string      `json:"description"`
	Bucket          string      `json:"bucket"`
	Repository      *Repository `json:"repository"`
	Binary          string      `json:"binary"`
	Facts           Facts       `json:"facts"`
	Review          *Ref        `json:"review"`
	Comments        string      `json:"comments"`
	Identities      []Ref       `json:"identities"`
	Tags            []Ref       `json:"tags"`
	BusinessService Ref         `json:"businessService"`
}

//
// With updates the resource using the model.
func (r *Application) With(m *model.Application) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Bucket = m.Bucket
	r.Comments = m.Comments
	r.Binary = m.Binary
	_ = json.Unmarshal(m.Repository, &r.Repository)
	_ = json.Unmarshal(m.Facts, &r.Facts)
	if m.Review != nil {
		ref := &Ref{}
		ref.With(m.Review.ID, "")
		r.Review = ref
	}
	r.BusinessService = r.ref(m.BusinessServiceID, m.BusinessService)
	r.Identities = []Ref{}
	for _, id := range m.Identities {
		ref := Ref{}
		ref.With(id.ID, id.Name)
		r.Identities = append(
			r.Identities,
			ref)
	}
	for _, tag := range m.Tags {
		ref := Ref{}
		ref.With(tag.ID, tag.Name)
		r.Tags = append(r.Tags, ref)
	}
}

//
// Model builds a model.
func (r *Application) Model() (m *model.Application) {
	m = &model.Application{
		Name:              r.Name,
		Description:       r.Description,
		Comments:          r.Comments,
		BusinessServiceID: r.BusinessService.ID,
		Binary:            r.Binary,
	}
	m.ID = r.ID
	if r.Repository != nil {
		m.Repository, _ = json.Marshal(r.Repository)
	}
	if r.Facts == nil {
		r.Facts = Facts{}
	}
	m.Facts, _ = json.Marshal(r.Facts)
	for _, ref := range r.Identities {
		m.Identities = append(
			m.Identities,
			model.Identity{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.Tags {
		m.Tags = append(
			m.Tags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}

	return
}

//
// Repository REST nested resource.
type Repository struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
	Path   string `json:"path"`
}

//
// Facts about the application.
type Facts map[string]interface{}
