package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// TagCategoryHandler handles the tag-type route.
type TagCategoryHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h TagCategoryHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tagcategories"))
	routeGroup.GET(api.TagCategoriesRoute, h.List)
	routeGroup.GET(api.TagCategoriesRoute+"/", h.List)
	routeGroup.POST(api.TagCategoriesRoute, h.Create)
	routeGroup.GET(api.TagCategoryRoute, h.Get)
	routeGroup.PUT(api.TagCategoryRoute, h.Update)
	routeGroup.DELETE(api.TagCategoryRoute, h.Delete)
	routeGroup.GET(api.TagCategoryTagsRoute, h.TagList)
	routeGroup.GET(api.TagCategoryTagsRoute+"/", h.TagList)
}

// Get godoc
// @summary Get a tag category by ID.
// @description Get a tag category by ID.
// @tags tagcategories
// @produce json
// @success 200 {object} api.TagCategory
// @router /tagcategories/{id} [get]
// @param id path int true "Tag Category ID"
func (h TagCategoryHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TagCategory{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := TagCategory{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all tag categories.
// @description List all tag categories.
// @tags tagcategories
// @produce json
// @success 200 {object} []api.TagCategory
// @router /tagcategories [get]
// @param name query string false "Optional category name filter"
func (h TagCategoryHandler) List(ctx *gin.Context) {
	var list []model.TagCategory
	db := h.preLoad(h.DB(ctx), clause.Associations)
	if name, found := ctx.GetQuery(Name); found {
		db = db.Where("name = ?", name)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []TagCategory{}
	for i := range list {
		r := TagCategory{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a tag category.
// @description Create a tag category.
// @tags tagcategories
// @accept json
// @produce json
// @success 201 {object} api.TagCategory
// @router /tagcategories [post]
// @param tag_type body api.TagCategory true "Tag Category data"
func (h TagCategoryHandler) Create(ctx *gin.Context) {
	r := TagCategory{}
	err := h.Bind(ctx, &r)
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
// @summary Delete a tag category.
// @description Delete a tag category.
// @tags tagcategories
// @success 204
// @router /tagcategories/{id} [delete]
// @param id path int true "Tag Category ID"
func (h TagCategoryHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TagCategory{}
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
// @summary Update a tag category.
// @description Update a tag category.
// @tags tagcategories
// @accept json
// @success 204
// @router /tagcategories/{id} [put]
// @param id path int true "Tag Category ID"
// @param tag_type body api.TagCategory true "Tag Category data"
func (h TagCategoryHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &TagCategory{}
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

// TagList godoc
// @summary List the tags in the tag category.
// @description List the tags in the tag category.
// @tags tagcategories
// @produce json
// @success 200 {object} []api.Tag
// @router /tagcategories/{id}/tags [get]
// @param id path int true "Tag Category ID"
// @param name query string false "Optional tag name filter"
func (h TagCategoryHandler) TagList(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TagCategory{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	var list []model.Tag
	db := h.DB(ctx)
	if name, found := ctx.GetQuery(Name); found {
		db = db.Where("name = ?", name)
	}
	result = db.Find(&list, "CategoryID = ?", id)
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

	ctx.JSON(http.StatusOK, resources)
}

// TagCategory REST resource.
type TagCategory = resource.TagCategory
