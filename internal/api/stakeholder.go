package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// StakeholderHandler handles stakeholder routes.
type StakeholderHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h StakeholderHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("stakeholders"), Transaction)
	routeGroup.GET(api.StakeholdersRoute, h.List)
	routeGroup.GET(api.StakeholdersRoute+"/", h.List)
	routeGroup.POST(api.StakeholdersRoute, h.Create)
	routeGroup.GET(api.StakeholderRoute, h.Get)
	routeGroup.PUT(api.StakeholderRoute, h.Update)
	routeGroup.DELETE(api.StakeholderRoute, h.Delete)
}

// Get godoc
// @summary Get a stakeholder by ID.
// @description Get a stakeholder by ID.
// @tags stakeholders
// @produce json
// @success 200 {object} api.Stakeholder
// @router /stakeholders/{id} [get]
// @param id path int true "Stakeholder ID"
func (h StakeholderHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Stakeholder{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := Stakeholder{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all stakeholders.
// @description List all stakeholders.
// @tags stakeholders
// @produce json
// @success 200 {object} []api.Stakeholder
// @router /stakeholders [get]
func (h StakeholderHandler) List(ctx *gin.Context) {
	var list []model.Stakeholder
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Stakeholder{}
	for i := range list {
		r := Stakeholder{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a stakeholder.
// @description Create a stakeholder.
// @tags stakeholders
// @accept json
// @produce json
// @success 201 {object} api.Stakeholder
// @router /stakeholders [post]
// @param stakeholder body api.Stakeholder true "Stakeholder data"
func (h StakeholderHandler) Create(ctx *gin.Context) {
	r := &Stakeholder{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Omit(clause.Associations).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.DB(ctx).Model(m).Association("Groups").Replace(m.Groups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("Owns").Replace(m.Owns)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("Contributes").Replace(m.Contributes)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("MigrationWaves").Replace(m.MigrationWaves)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a stakeholder.
// @description Delete a stakeholder.
// @tags stakeholders
// @success 204
// @router /stakeholders/{id} [delete]
// @param id path int true "Stakeholder ID"
func (h StakeholderHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Stakeholder{}
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
// @summary Update a stakeholder.
// @description Update a stakeholder.
// @tags stakeholders
// @accept json
// @success 204
// @router /stakeholders/{id} [put]
// @param id path int true "Stakeholder ID"
// @param stakeholder body api.Stakeholder true "Stakeholder data"
func (h StakeholderHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Stakeholder{}
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
	db = h.DB(ctx).Model(m)
	err = db.Association("Groups").Replace(m.Groups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Owns").Replace(m.Owns)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Contributes").Replace(m.Contributes)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("MigrationWaves").Replace(m.MigrationWaves)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// Stakeholder REST resource.
type Stakeholder = resource.Stakeholder
