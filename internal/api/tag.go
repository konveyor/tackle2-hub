package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// TagHandler handles tag routes.
type TagHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h TagHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tags"))
	routeGroup.GET(api.TagsRoute, h.List)
	routeGroup.GET(api.TagsRoute+"/", h.List)
	routeGroup.POST(api.TagsRoute, h.Create)
	routeGroup.GET(api.TagRoute, h.Get)
	routeGroup.PUT(api.TagRoute, h.Update)
	routeGroup.DELETE(api.TagRoute, h.Delete)
}

// Get godoc
// @summary Get a tag by ID.
// @description Get a tag by ID.
// @tags tags
// @produce json
// @success 200 {object} api.Tag
// @router /tags/{id} [get]
// @param id path int true "Tag ID"
func (h TagHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tag{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := Tag{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all tags.
// @description List all tags.
// @tags tags
// @produce json
// @success 200 {object} []api.Tag
// @router /tags [get]
func (h TagHandler) List(ctx *gin.Context) {
	var list []model.Tag
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Tag{}
	for i := range list {
		r := Tag{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a tag.
// @description Create a tag.
// @tags tags
// @accept json
// @produce json
// @success 201 {object} api.Tag
// @router /tags [post]
// @param tag body Tag true "Tag data"
func (h TagHandler) Create(ctx *gin.Context) {
	r := &Tag{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a tag.
// @description Delete a tag.
// @tags tags
// @success 204
// @router /tags/{id} [delete]
// @param id path int true "Tag ID"
func (h TagHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tag{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a tag.
// @description Update a tag.
// @tags tags
// @accept json
// @success 204
// @router /tags/{id} [put]
// @param id path int true "Tag ID"
// @param tag body api.Tag true "Tag data"
func (h TagHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Tag{}
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

// Tag REST resource.
type Tag = resource.Tag
