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
	StakeholdersRoot = "/stakeholders"
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
// @router /stakeholders/{id} [get]
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
// @router /stakeholders [get]
func (h StakeholderHandler) List(ctx *gin.Context) {
	var list []model.Stakeholder
	db := h.preLoad(
		h.DB,
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

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a stakeholder.
// @description Create a stakeholder.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Stakeholder
// @router /stakeholders [post]
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
// @router /stakeholders/{id} [delete]
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
// @router /stakeholders/{id} [put]
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
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Groups           []Ref  `json:"stakeholderGroups"`
	BusinessServices []Ref  `json:"businessServices"`
	JobFunction      *Ref   `json:"jobFunction"`
}

//
// With updates the resource with the model.
func (r *Stakeholder) With(m *model.Stakeholder) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Email = m.Email
	if m.JobFunction != nil {
		ref := &Ref{}
		ref.With(m.JobFunction.ID, m.JobFunction.Name)
		r.JobFunction = ref
	}
	for _, g := range m.Groups {
		ref := Ref{}
		ref.With(g.ID, g.Name)
		r.Groups = append(r.Groups, ref)
	}
	for _, b := range m.BusinessServices {
		ref := Ref{}
		ref.With(b.ID, b.Name)
		r.BusinessServices = append(r.BusinessServices, ref)
	}
}

//
// Model builds a model.
func (r *Stakeholder) Model() (m *model.Stakeholder) {
	m = &model.Stakeholder{
		Name:  r.Name,
		Email: r.Email,
	}
	m.ID = r.ID
	if r.JobFunction != nil {
		m.JobFunctionID = &r.JobFunction.ID
	}
	for _, g := range r.Groups {
		m.Groups = append(m.Groups, model.StakeholderGroup{Model: model.Model{ID: g.ID}})
	}
	for _, b := range r.BusinessServices {
		m.BusinessServices = append(m.BusinessServices, model.BusinessService{Model: model.Model{ID: b.ID}})
	}
	return
}
