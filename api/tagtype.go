package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
)

//
// Routes
const (
	TagTypesRoot = "/tagtypes"
	TagTypeRoot  = TagTypesRoot + "/:" + ID
)

//
// TagTypeHandler handles the tag-type route.
type TagTypeHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h TagTypeHandler) AddRoutes(e *gin.Engine) {
	e.GET(TagTypesRoot, h.List)
	e.GET(TagTypesRoot+"/", h.List)
	e.POST(TagTypesRoot, h.Create)
	e.GET(TagTypeRoot, h.Get)
	e.PUT(TagTypeRoot, h.Update)
	e.DELETE(TagTypeRoot, h.Delete)
}

// Get godoc
// @summary Get a tag type by ID.
// @description Get a tag type by ID.
// @tags get
// @produce json
// @success 200 {object} api.TagType
// @router /tagtypes/{id} [get]
// @param id path string true "Tag Type ID"
func (h TagTypeHandler) Get(ctx *gin.Context) {
	m := &model.TagType{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB, "Tags")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	resource := TagType{}
	resource.With(m)
	ctx.JSON(http.StatusOK, resource)
}

// List godoc
// @summary List all tag types.
// @description List all tag types.
// @tags get
// @produce json
// @success 200 {object} []api.TagType
// @router /tagtypes [get]
func (h TagTypeHandler) List(ctx *gin.Context) {
	var list []model.TagType
	db := h.preLoad(h.DB, "Tags")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []TagType{}
	for i := range list {
		r := TagType{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a tag type.
// @description Create a tag type.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.TagType
// @router /tagtypes [post]
// @param tag_type body api.TagType true "Tag Type data"
func (h TagTypeHandler) Create(ctx *gin.Context) {
	r := TagType{}
	err := ctx.BindJSON(&r)
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
	r.With(m)

	ctx.JSON(http.StatusCreated, m)
}

// Delete godoc
// @summary Delete a tag type.
// @description Delete a tag type.
// @tags delete
// @success 204
// @router /tagtypes/{id} [delete]
// @param id path string true "Tag Type ID"
func (h TagTypeHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	result := h.DB.Delete(&model.TagType{}, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a tag type.
// @description Update a tag type.
// @tags update
// @accept json
// @success 204
// @router /tagtypes/{id} [put]
// @param id path string true "Tag Type ID"
// @param tag_type body api.TagType true "Tag Type data"
func (h TagTypeHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	r := &TagType{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Model(&TagType{}).Where("id = ?", id).Omit("id").Updates(m)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// TagType REST resource.
type TagType struct {
	Resource
	Name     string `json:"name" binding:"required"`
	Username string `json:"username"`
	Rank     uint   `json:"rank"`
	Color    string `json:"colour"`
	Tags     []Ref  `json:"tags"`
}

//
// With updates the resource with the model.
func (r *TagType) With(m *model.TagType) {
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
func (r *TagType) Model() (m *model.TagType) {
	m = &model.TagType{
		Name:     r.Name,
		Username: r.Username,
		Rank:     r.Rank,
		Color:    r.Color,
	}
	m.ID = r.ID
	return
}
