package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	TagCategoriesRoot = "/tagcategories"
	TagCategoryRoot   = TagCategoriesRoot + "/:" + ID
)

//
// TagCategoryHandler handles the tag-type route.
type TagCategoryHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h TagCategoryHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("tagcategories"))
	routeGroup.GET(TagCategoriesRoot, h.List)
	routeGroup.GET(TagCategoriesRoot+"/", h.List)
	routeGroup.POST(TagCategoriesRoot, h.Create)
	routeGroup.GET(TagCategoryRoot, h.Get)
	routeGroup.PUT(TagCategoryRoot, h.Update)
	routeGroup.DELETE(TagCategoryRoot, h.Delete)
}

// Get godoc
// @summary Get a tag category by ID.
// @description Get a tag category by ID.
// @tags get
// @produce json
// @success 200 {object} api.TagCategory
// @router /tagcategories/{id} [get]
// @param id path string true "Tag Category ID"
func (h TagCategoryHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TagCategory{}
	db := h.preLoad(h.DB, clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}

	resource := TagCategory{}
	resource.With(m)
	ctx.JSON(http.StatusOK, resource)
}

// List godoc
// @summary List all tag categories.
// @description List all tag categories.
// @tags get
// @produce json
// @success 200 {object} []api.TagCategory
// @router /tagcategories [get]
func (h TagCategoryHandler) List(ctx *gin.Context) {
	var list []model.TagCategory
	db := h.preLoad(h.DB, clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}
	resources := []TagCategory{}
	for i := range list {
		r := TagCategory{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a tag category.
// @description Create a tag category.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.TagCategory
// @router /tagcategories [post]
// @param tag_type body api.TagCategory true "Tag Category data"
func (h TagCategoryHandler) Create(ctx *gin.Context) {
	r := TagCategory{}
	err := ctx.BindJSON(&r)
	if err != nil {
		h.reportError(ctx, err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB.Create(m)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a tag category.
// @description Delete a tag category.
// @tags delete
// @success 204
// @router /tagcategories/{id} [delete]
// @param id path string true "Tag Category ID"
func (h TagCategoryHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TagCategory{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}
	result = h.DB.Delete(m)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a tag category.
// @description Update a tag category.
// @tags update
// @accept json
// @success 204
// @router /tagcategories/{id} [put]
// @param id path string true "Tag Category ID"
// @param tag_type body api.TagCategory true "Tag Category data"
func (h TagCategoryHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &TagCategory{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.reportError(ctx, err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB.Model(&model.TagCategory{Model: model.Model{ID: id}})
	db = db.Omit(clause.Associations)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// TagCategory REST resource.
type TagCategory struct {
	Resource
	Name     string `json:"name" binding:"required"`
	Username string `json:"username"`
	Rank     uint   `json:"rank"`
	Color    string `json:"colour"`
	Tags     []Ref  `json:"tags"`
}

//
// With updates the resource with the model.
func (r *TagCategory) With(m *model.TagCategory) {
	r.Resource.With(&m.Model)
	r.ID = m.ID
	r.Name = m.Name
	r.Username = m.Username
	r.Rank = m.Rank
	r.Color = m.Color
	for _, tag := range m.Tags {
		ref := Ref{}
		ref.With(tag.ID, tag.Name)
		r.Tags = append(r.Tags, ref)
	}
}

//
// Model builds a model.
func (r *TagCategory) Model() (m *model.TagCategory) {
	m = &model.TagCategory{
		Name:     r.Name,
		Username: r.Username,
		Rank:     r.Rank,
		Color:    r.Color,
	}
	m.ID = r.ID
	return
}
