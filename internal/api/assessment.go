package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// AssessmentHandler handles Assessment resource routes.
type AssessmentHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AssessmentHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("assessments"), Transaction)
	routeGroup.GET(api.AssessmentsRoute, h.List)
	routeGroup.GET(api.AssessmentsRoute+"/", h.List)
	routeGroup.GET(api.AssessmentRoute, h.Get)
	routeGroup.PUT(api.AssessmentRoute, h.Update)
	routeGroup.DELETE(api.AssessmentRoute, h.Delete)
}

// Get godoc
// @summary Get an assessment by ID.
// @description Get an assessment by ID.
// @tags questionnaires
// @produce json
// @success 200 {object} api.Assessment
// @router /assessments/{id} [get]
// @param id path int true "Assessment ID"
func (h AssessmentHandler) Get(ctx *gin.Context) {
	m := &model.Assessment{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Assessment{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all assessments.
// @description List all assessments.
// @tags assessments
// @produce json
// @success 200 {object} []api.Assessment
// @router /assessments [get]
func (h AssessmentHandler) List(ctx *gin.Context) {
	var list []model.Assessment
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Assessment{}
	for i := range list {
		r := Assessment{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Delete godoc
// @summary Delete an assessment.
// @description Delete an assessment.
// @tags assessments
// @success 204
// @router /assessments/{id} [delete]
// @param id path int true "Assessment ID"
func (h AssessmentHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Assessment{}
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
// @summary Update an assessment.
// @description Update an assessment.
// @tags assessments
// @accept json
// @success 204
// @router /assessments/{id} [put]
// @param id path int true "Assessment ID"
// @param assessment body api.Assessment true "Assessment data"
func (h AssessmentHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Assessment{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations, "Thresholds", "RiskMessages")
	result := db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
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

// Assessment REST resource.
type Assessment = resource.Assessment
