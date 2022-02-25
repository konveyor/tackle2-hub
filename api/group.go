package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
	"strconv"
)

//
// Routes
const (
	StakeholderGroupsRoot = "/stakeholdergroups"
	StakeholderGroupRoot  = StakeholderGroupsRoot + "/:" + ID
)

//
// StakeholderGroupHandler handles stakeholder group routes.
type StakeholderGroupHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h StakeholderGroupHandler) AddRoutes(e *gin.Engine) {
	e.GET(StakeholderGroupsRoot, h.List)
	e.GET(StakeholderGroupsRoot+"/", h.List)
	e.POST(StakeholderGroupsRoot, h.Create)
	e.GET(StakeholderGroupRoot, h.Get)
	e.PUT(StakeholderGroupRoot, h.Update)
	e.DELETE(StakeholderGroupRoot, h.Delete)
}

// Get godoc
// @summary Get a stakeholder group by ID.
// @description Get a stakeholder group by ID.
// @tags get
// @produce json
// @success 200 {object} api.StakeholderGroup
// @router /stakeholdergroups/{id} [get]
// @param id path string true "Stakeholder Group ID"
func (h StakeholderGroupHandler) Get(ctx *gin.Context) {
	m := &model.StakeholderGroup{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB, "Stakeholders")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := StakeholderGroup{}
	r.With(m)

	ctx.JSON(http.StatusOK, m)
}

// List godoc
// @summary List all stakeholder groups.
// @description List all stakeholder groups.
// @tags get
// @produce json
// @success 200 {object} []api.StakeholderGroup
// @router /stakeholdergroups [get]
func (h StakeholderGroupHandler) List(ctx *gin.Context) {
	var list []model.StakeholderGroup
	db := h.preLoad(h.DB, "Stakeholders")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []StakeholderGroup{}
	for i := range list {
		r := StakeholderGroup{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a stakeholder group.
// @description Create a stakeholder group.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.StakeholderGroup
// @router /stakeholdergroups [post]
// @param stakeholder_group body api.StakeholderGroup true "Stakeholder Group data"
func (h StakeholderGroupHandler) Create(ctx *gin.Context) {
	r := &StakeholderGroup{}
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
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a stakeholder group.
// @description Delete a stakeholder group.
// @tags delete
// @success 204
// @router /stakeholdergroups/{id} [delete]
// @param id path string true "Stakeholder Group ID"
func (h StakeholderGroupHandler) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param(ID))
	m := &model.StakeholderGroup{}
	m.ID = uint(id)
	result := h.DB.Select("Stakeholders").Delete(m)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a stakeholder group.
// @description Update a stakeholder group.
// @tags update
// @accept json
// @success 204
// @router /stakeholdergroups/{id} [put]
// @param id path string true "Stakeholder Group ID"
// @param stakeholder_group body api.StakeholderGroup true "Stakeholder Group data"
func (h StakeholderGroupHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	r := &StakeholderGroup{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Model(&model.StakeholderGroup{}).Where("id = ?", id).Omit("id").Updates(m)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	err = h.DB.Model(m).Association("Stakeholders").Replace("Stakeholders", m.Stakeholders)
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// StakeholderGroup REST resource.
type StakeholderGroup struct {
	Resource
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Stakeholders []Ref  `json:"stakeholders"`
}

//
// With updates the resource with the model.
func (r *StakeholderGroup) With(m *model.StakeholderGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
}

//
// Model builds a model.
func (r *StakeholderGroup) Model() (m *model.StakeholderGroup) {
	m = &model.StakeholderGroup{
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	for _, s := range r.Stakeholders {
		m.Stakeholders = append(m.Stakeholders, model.Stakeholder{Model: model.Model{ID: s.ID}})
	}
	return
}
