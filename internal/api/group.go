package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// StakeholderGroupHandler handles stakeholder group routes.
type StakeholderGroupHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h StakeholderGroupHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("stakeholdergroups"), Transaction)
	routeGroup.GET(api.StakeholderGroupsRoute, h.List)
	routeGroup.GET(api.StakeholderGroupsRoute+"/", h.List)
	routeGroup.POST(api.StakeholderGroupsRoute, h.Create)
	routeGroup.GET(api.StakeholderGroupRoute, h.Get)
	routeGroup.PUT(api.StakeholderGroupRoute, h.Update)
	routeGroup.DELETE(api.StakeholderGroupRoute, h.Delete)
}

// Get godoc
// @summary Get a stakeholder group by ID.
// @description Get a stakeholder group by ID.
// @tags stakeholdergroups
// @produce json
// @success 200 {object} api.StakeholderGroup
// @router /stakeholdergroups/{id} [get]
// @param id path int true "Stakeholder Group ID"
func (h StakeholderGroupHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.StakeholderGroup{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := StakeholderGroup{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all stakeholder groups.
// @description List all stakeholder groups.
// @tags stakeholdergroups
// @produce json
// @success 200 {object} []api.StakeholderGroup
// @router /stakeholdergroups [get]
func (h StakeholderGroupHandler) List(ctx *gin.Context) {
	var list []model.StakeholderGroup
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []StakeholderGroup{}
	for i := range list {
		r := StakeholderGroup{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a stakeholder group.
// @description Create a stakeholder group.
// @tags stakeholdergroups
// @accept json
// @produce json
// @success 201 {object} api.StakeholderGroup
// @router /stakeholdergroups [post]
// @param stakeholder_group body api.StakeholderGroup true "Stakeholder Group data"
func (h StakeholderGroupHandler) Create(ctx *gin.Context) {
	r := &StakeholderGroup{}
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
	err = h.DB(ctx).Model(m).Association("Stakeholders").Replace(m.Stakeholders)
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
// @summary Delete a stakeholder group.
// @description Delete a stakeholder group.
// @tags stakeholdergroups
// @success 204
// @router /stakeholdergroups/{id} [delete]
// @param id path int true "Stakeholder Group ID"
func (h StakeholderGroupHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.StakeholderGroup{}
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
// @summary Update a stakeholder group.
// @description Update a stakeholder group.
// @tags stakeholdergroups
// @accept json
// @success 204
// @router /stakeholdergroups/{id} [put]
// @param id path int true "Stakeholder Group ID"
// @param stakeholder_group body api.StakeholderGroup true "Stakeholder Group data"
func (h StakeholderGroupHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &StakeholderGroup{}
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
	err = db.Association("Stakeholders").Replace(m.Stakeholders)
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

// StakeholderGroup REST resource.
type StakeholderGroup = resource.StakeholderGroup
