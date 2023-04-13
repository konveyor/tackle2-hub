package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	TagsRoot = "/tags"
	TagRoot  = TagsRoot + "/:" + ID
)

//
// TagHandler handles tag routes.
type TagHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h TagHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tags"))
	routeGroup.GET(TagsRoot, h.List)
	routeGroup.GET(TagsRoot+"/", h.List)
	routeGroup.POST(TagsRoot, h.Create)
	routeGroup.GET(TagRoot, h.Get)
	routeGroup.PUT(TagRoot, h.Update)
	routeGroup.DELETE(TagRoot, h.Delete)
}

// Get godoc
// @summary Get a tag by ID.
// @description Get a tag by ID.
// @tags tags
// @produce json
// @success 200 {object} api.Tag
// @router /tags/{id} [get]
// @param id path string true "Tag ID"
func (h TagHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tag{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	resource := Tag{}
	resource.With(m)
	h.Render(ctx, http.StatusOK, resource)
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
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
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

	h.Render(ctx, http.StatusOK, resources)
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

	h.Render(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a tag.
// @description Delete a tag.
// @tags tags
// @success 204
// @router /tags/{id} [delete]
// @param id path string true "Tag ID"
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

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a tag.
// @description Update a tag.
// @tags tags
// @accept json
// @success 204
// @router /tags/{id} [put]
// @param id path string true "Tag ID"
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
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Tag REST resource.
type Tag struct {
	Resource
	Name     string `json:"name" binding:"required"`
	Category Ref    `json:"category" binding:"required"`
}

//
// With updates the resource with the model.
func (r *Tag) With(m *model.Tag) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Category = r.ref(m.CategoryID, &m.Category)
}

//
// Model builds a model.
func (r *Tag) Model() (m *model.Tag) {
	m = &model.Tag{
		Name:       r.Name,
		CategoryID: r.Category.ID,
	}
	m.ID = r.ID
	return
}
