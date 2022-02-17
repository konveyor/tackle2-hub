package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
	"strconv"
)

//
// Kind
const (
	StakeholderKind = "stakeholder"
)

//
// Routes
const (
	StakeholdersRoot = ControlsRoot + "/stakeholder"
	StakeholderRoot  = StakeholdersRoot + "/:" + ID
)

//
// StakeholderHandler handles stakeholder routes.
type StakeholderHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h StakeholderHandler) AddRoutes(e *gin.Engine) {
	e.GET(StakeholdersRoot, h.List)
	e.GET(StakeholdersRoot+"/", h.List)
	e.POST(StakeholdersRoot, h.Create)
	e.GET(StakeholderRoot, h.Get)
	e.PUT(StakeholderRoot, h.Update)
	e.DELETE(StakeholderRoot, h.Delete)
}

// Get godoc
// @summary Get a stakeholder by ID.
// @description Get a stakeholder by ID.
// @tags get
// @produce json
// @success 200 {object} api.Stakeholder
// @router /controls/stakeholder/{id} [get]
// @param id path string true "Stakeholder ID"
func (h StakeholderHandler) Get(ctx *gin.Context) {
	m := &model.Stakeholder{}
	id := ctx.Param(ID)
	db := h.preLoad(
		h.DB,
		"JobFunction",
		"BusinessServices",
		"StakeholderGroups")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	resource := Stakeholder{}
	resource.With(m)
	ctx.JSON(http.StatusOK, resource)
}

// List godoc
// @summary List all stakeholders.
// @description List all stakeholders.
// @tags get
// @produce json
// @success 200 {object} []api.Stakeholder
// @router /controls/stakeholder [get]
func (h StakeholderHandler) List(ctx *gin.Context) {
	var count int64
	var list []model.Stakeholder
	h.DB.Model(model.Stakeholder{}).Count(&count)
	pagination := NewPagination(ctx)
	db := pagination.apply(h.DB)
	db = h.preLoad(
		db,
		"JobFunction",
		"BusinessServices",
		"Groups")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Stakeholder{}
	for i := range list {
		r := Stakeholder{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.listResponse(ctx, StakeholderKind, resources, int(count))
}

// Create godoc
// @summary Create a stakeholder.
// @description Create a stakeholder.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Stakeholder
// @router /controls/stakeholder [post]
// @param stakeholder body api.Stakeholder true "Stakeholder data"
func (h StakeholderHandler) Create(ctx *gin.Context) {
	r := &Stakeholder{}
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
// @summary Delete a stakeholder.
// @description Delete a stakeholder.
// @tags delete
// @success 204
// @router /controls/stakeholder/{id} [delete]
// @param id path string true "Stakeholder ID"
func (h StakeholderHandler) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param(ID))
	m := &model.Stakeholder{}
	m.ID = uint(id)
	result := h.DB.Select("Groups").Delete(m)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a stakeholder.
// @description Update a stakeholder.
// @tags update
// @accept json
// @success 204
// @router /controls/stakeholder/{id} [put]
// @param id path string true "Stakeholder ID"
// @param stakeholder body api.Stakeholder true "Stakeholder data"
func (h StakeholderHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	resource := Stakeholder{}
	err := ctx.BindJSON(&resource)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	updates := resource.Model()
	result := h.DB.Model(&model.Stakeholder{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	err = h.DB.Model(updates).Association("Groups").Replace("Groups", updates.Groups)
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Stakeholder REST resource.
type Stakeholder struct {
	Resource
	DisplayName      string             `json:"displayName" binding:"required"`
	Email            string             `json:"email" binding:"required"`
	Groups           []StakeholderGroup `json:"stakeholderGroups"`
	BusinessServices []BusinessService  `json:"businessServices"`
	JobFunction      struct {
		ID   *uint  `json:"id"`
		Role string `json:"role"`
	} `json:"jobFunction"`
}

//
// With updates the resource with the model.
func (r *Stakeholder) With(m *model.Stakeholder) {
	r.Resource.With(&m.Model)
	r.DisplayName = m.DisplayName
	r.Email = m.Email
	r.JobFunction.ID = m.JobFunctionID
	if m.JobFunction != nil {
		r.JobFunction.Role = m.JobFunction.Role
	}
	for _, g := range m.Groups {
		group := StakeholderGroup{}
		group.With(&g)
		r.Groups = append(r.Groups, group)
	}
	for _, b := range m.BusinessServices {
		business := BusinessService{}
		business.With(&b)
		r.BusinessServices = append(r.BusinessServices, business)
	}
}

//
// Model builds a model.
func (r *Stakeholder) Model() (m *model.Stakeholder) {
	m = &model.Stakeholder{
		DisplayName:   r.DisplayName,
		Email:         r.Email,
		JobFunctionID: r.JobFunction.ID,
	}
	m.ID = r.ID
	for _, g := range r.Groups {
		m.Groups = append(m.Groups, *g.Model())
	}
	for _, b := range r.BusinessServices {
		m.BusinessServices = append(m.BusinessServices, *b.Model())
	}
	return
}
