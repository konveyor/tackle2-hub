package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// MigrationWaveHandler handles Migration Wave resource routes.
type MigrationWaveHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h MigrationWaveHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("migrationwaves"), Transaction)
	routeGroup.GET(api.MigrationWavesRoute, h.List)
	routeGroup.GET(api.MigrationWavesRoute+"/", h.List)
	routeGroup.GET(api.MigrationWaveRoute, h.Get)
	routeGroup.POST(api.MigrationWavesRoute, h.Create)
	routeGroup.DELETE(api.MigrationWaveRoute, h.Delete)
	routeGroup.PUT(api.MigrationWaveRoute, h.Update)
}

// Get godoc
// @summary Get a migration wave by ID.
// @description Get a migration wave by ID.
// @tags migrationwaves
// @produce json
// @success 200 {object} api.MigrationWave
// @router /migrationwaves/{id} [get]
// @param id path int true "Migration Wave ID"
func (h MigrationWaveHandler) Get(ctx *gin.Context) {
	m := &model.MigrationWave{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := MigrationWave{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all migration waves.
// @description List all migration waves.
// @tags migrationwaves
// @produce json
// @success 200 {object} []api.MigrationWave
// @router /migrationwaves [get]
func (h MigrationWaveHandler) List(ctx *gin.Context) {
	var list []model.MigrationWave
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []MigrationWave{}
	for i := range list {
		r := MigrationWave{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a migration wave.
// @description Create a migration wave.
// @tags migrationwaves
// @accept json
// @produce json
// @success 201 {object} api.MigrationWave
// @router /migrationwaves [post]
// @param migrationwave body api.MigrationWave true "Migration Wave data"
func (h MigrationWaveHandler) Create(ctx *gin.Context) {
	r := &MigrationWave{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	result := h.DB(ctx).Omit(clause.Associations).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.DB(ctx).Model(m).Association("Applications").Replace("Applications", m.Applications)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("Stakeholders").Replace("Stakeholders", m.Stakeholders)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("StakeholderGroups").Replace("StakeholderGroups", m.StakeholderGroups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Update godoc
// @summary Update a migration wave.
// @description Update a migration wave.
// @tags migrationwaves
// @accept json
// @success 204
// @router /migrationwaves/{id} [put]
// @param id path int true "MigrationWave id"
// @param migrationWave body api.MigrationWave true "MigrationWave data"
func (h MigrationWaveHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &MigrationWave{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result := db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.DB(ctx).Model(m).Association("Applications").Replace("Applications", m.Applications)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("Stakeholders").Replace("Stakeholders", m.Stakeholders)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("StakeholderGroups").Replace("StakeholderGroups", m.StakeholderGroups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Delete godoc
// @summary Delete a migration wave.
// @description Delete a migration wave.
// @tags migrationwaves
// @success 204
// @router /migrationwaves/{id} [delete]
// @param id path int true "MigrationWave id"
func (h MigrationWaveHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.MigrationWave{}
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

// MigrationWave REST Resource
type MigrationWave = resource.MigrationWave
