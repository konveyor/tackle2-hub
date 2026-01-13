package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// BusinessServiceHandler handles business-service routes.
type BusinessServiceHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h BusinessServiceHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("businessservices"))
	routeGroup.GET(api.BusinessServicesRoute, h.List)
	routeGroup.GET(api.BusinessServicesRoute+"/", h.List)
	routeGroup.POST(api.BusinessServicesRoute, h.Create)
	routeGroup.GET(api.BusinessServiceRoute, h.Get)
	routeGroup.PUT(api.BusinessServiceRoute, h.Update)
	routeGroup.DELETE(api.BusinessServiceRoute, h.Delete)
}

// Get godoc
// @summary Get a business service by ID.
// @description Get a business service by ID.
// @tags businessservices
// @produce json
// @success 200 {object} api.BusinessService
// @router /businessservices/{id} [get]
// @param id path int true "Business Service ID"
func (h BusinessServiceHandler) Get(ctx *gin.Context) {
	m := &model.BusinessService{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := BusinessService{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all business services.
// @description List all business services.
// @tags businessservices
// @produce json
// @success 200 {object} api.BusinessService
// @router /businessservices [get]
func (h BusinessServiceHandler) List(ctx *gin.Context) {
	var list []model.BusinessService
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []BusinessService{}
	for i := range list {
		r := BusinessService{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a business service.
// @description Create a business service.
// @tags businessservices
// @accept json
// @produce json
// @success 201 {object} api.BusinessService
// @router /businessservices [post]
// @param business_service body api.BusinessService true "Business service data"
func (h BusinessServiceHandler) Create(ctx *gin.Context) {
	r := &BusinessService{}
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

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a business service.
// @description Delete a business service.
// @tags businessservices
// @success 204
// @router /businessservices/{id} [delete]
// @param id path int true "Business service ID"
func (h BusinessServiceHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.BusinessService{}
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
// @summary Update a business service.
// @description Update a business service.
// @tags businessservices
// @accept json
// @success 204
// @router /businessservices/{id} [put]
// @param id path int true "Business service ID"
// @param business_service body api.BusinessService true "Business service data"
func (h BusinessServiceHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &BusinessService{}
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

	h.Status(ctx, http.StatusNoContent)
}

// BusinessService REST resource.
type BusinessService = resource.BusinessService
