package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// ReviewHandler handles review routes.
type ReviewHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h ReviewHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("reviews"))
	routeGroup.GET(api.ReviewsRoute, h.List)
	routeGroup.GET(api.ReviewsRoute+"/", h.List)
	routeGroup.POST(api.ReviewsRoute, h.Create)
	routeGroup.GET(api.ReviewRoute, h.Get)
	routeGroup.PUT(api.ReviewRoute, h.Update)
	routeGroup.DELETE(api.ReviewRoute, h.Delete)
	routeGroup.POST(api.CopyRoute, h.CopyReview, Transaction)
}

// Get godoc
// @summary Get a review by ID.
// @description Get a review by ID.
// @tags reviews
// @produce json
// @success 200 {object} api.Review
// @router /reviews/{id} [get]
// @param id path int true "Review ID"
func (h ReviewHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Review{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Review{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all reviews.
// @description List all reviews.
// @tags reviews
// @produce json
// @success 200 {object} []api.Review
// @router /reviews [get]
func (h ReviewHandler) List(ctx *gin.Context) {
	var list []model.Review
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Review{}
	for i := range list {
		r := Review{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a review.
// @description Create a review.
// @tags reviews
// @accept json
// @produce json
// @success 201 {object} api.Review
// @router /reviews [post]
// @param review body api.Review true "Review data"
func (h ReviewHandler) Create(ctx *gin.Context) {
	review := Review{}
	err := h.Bind(ctx, &review)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := review.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	review.With(m)

	h.Respond(ctx, http.StatusCreated, review)
}

// Delete godoc
// @summary Delete a review.
// @description Delete a review.
// @tags reviews
// @success 204
// @router /reviews/{id} [delete]
// @param id path int true "Review ID"
func (h ReviewHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Review{}
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
// @summary Update a review.
// @description Update a review.
// @tags reviews
// @accept json
// @success 204
// @router /reviews/{id} [put]
// @param id path int true "Review ID"
// @param review body api.Review true "Review data"
func (h ReviewHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Review{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db.Omit(clause.Associations)
	result := db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// CopyReview godoc
// @summary Copy a review from one application to others.
// @description Copy a review from one application to others.
// @tags reviews
// @accept json
// @success 204
// @router /reviews/copy [post]
// @param copy_request body api.CopyRequest true "Review copy request data"
func (h ReviewHandler) CopyReview(ctx *gin.Context) {
	c := CopyRequest{}
	err := h.Bind(ctx, &c)
	if err != nil {
		return
	}

	m := model.Review{}
	result := h.DB(ctx).First(&m, c.SourceReview)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for _, id := range c.TargetApplications {
		copied := &model.Review{
			BusinessCriticality: m.BusinessCriticality,
			EffortEstimate:      m.EffortEstimate,
			ProposedAction:      m.ProposedAction,
			WorkPriority:        m.WorkPriority,
			Comments:            m.Comments,
			ApplicationID:       &id,
		}
		result = h.DB(ctx).Delete(&model.Review{}, "applicationid = ?", id)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
		result = h.DB(ctx).Create(copied)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
	}
	h.Status(ctx, http.StatusNoContent)
}

// Review REST resource.
type Review = resource.Review

// CopyRequest REST resource.
type CopyRequest struct {
	SourceReview       uint   `json:"sourceReview" binding:"required"`
	TargetApplications []uint `json:"targetApplications" binding:"required"`
}
