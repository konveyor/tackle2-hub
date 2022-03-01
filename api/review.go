package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
)

//
// Routes
const (
	ReviewsRoot = "/reviews"
	ReviewRoot  = ReviewsRoot + "/:" + ID
	CopyRoot    = ReviewsRoot + "/copy"
)

//
// ReviewHandler handles review routes.
type ReviewHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h ReviewHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, "reviews"))
	routeGroup.GET(ReviewsRoot, h.List)
	routeGroup.GET(ReviewsRoot+"/", h.List)
	routeGroup.POST(ReviewsRoot, h.Create)
	routeGroup.GET(ReviewRoot, h.Get)
	routeGroup.PUT(ReviewRoot, h.Update)
	routeGroup.DELETE(ReviewRoot, h.Delete)
	routeGroup.POST(CopyRoot, h.CopyReview)
}

// Get godoc
// @summary Get a review by ID.
// @description Get a review by ID.
// @tags get
// @produce json
// @success 200 {object} []api.Review
// @router /reviews/{id} [get]
// @param id path string true "Review ID"
func (h ReviewHandler) Get(ctx *gin.Context) {
	m := &model.Review{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB, "Application")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Review{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all reviews.
// @description List all reviews.
// @tags get
// @produce json
// @success 200 {object} []api.Review
// @router /reviews [get]
func (h ReviewHandler) List(ctx *gin.Context) {
	var list []model.Review
	db := h.preLoad(h.DB, "Application")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Review{}
	for i := range list {
		r := Review{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a review.
// @description Create a review.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Review
// @router /reviews [post]
// @param review body api.Review true "Review data"
func (h ReviewHandler) Create(ctx *gin.Context) {
	review := Review{}
	err := ctx.BindJSON(&review)
	if err != nil {
		return
	}
	m := review.Model()
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	review.With(m)

	ctx.JSON(http.StatusCreated, review)
}

// Delete godoc
// @summary Delete a review.
// @description Delete a review.
// @tags delete
// @success 204
// @router /reviews/{id} [delete]
// @param id path string true "Review ID"
func (h ReviewHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	result := h.DB.Delete(&Review{}, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a review.
// @description Update a review.
// @tags update
// @accept json
// @success 204
// @router /reviews/{id} [put]
// @param id path string true "Review ID"
// @param review body api.Review true "Review data"
func (h ReviewHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	updates := Review{}
	err := ctx.BindJSON(&updates)
	if err != nil {
		return
	}
	result := h.DB.Model(&Review{}).Where("id = ?", id).Omit("id").Updates(updates)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// CopyReview godoc
// @summary Copy a review from one application to others.
// @description Copy a review from one application to others.
// @tags copy
// @accept json
// @success 204
// @router /reviews/copy [post]
// @param copy_request body api.CopyRequest true "Review copy request data"
func (h ReviewHandler) CopyReview(ctx *gin.Context) {
	c := CopyRequest{}
	err := ctx.BindJSON(&c)
	if err != nil {
		return
	}

	m := model.Review{}
	result := h.DB.First(&m, c.SourceReview)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	for _, id := range c.TargetApplications {
		copied := model.Review{
			BusinessCriticality: m.BusinessCriticality,
			EffortEstimate:      m.EffortEstimate,
			ProposedAction:      m.ProposedAction,
			WorkPriority:        m.WorkPriority,
			Comments:            m.Comments,
			ApplicationID:       id,
		}
		existing := []model.Review{}
		result = h.DB.Find(&existing, "applicationid = ?", id)
		if result.Error != nil {
			h.createFailed(ctx, result.Error)
			return
		}
		// if the application doesn't already have a review, create one.
		if len(existing) == 0 {
			result = h.DB.Create(&copied)
			if result.Error != nil {
				h.createFailed(ctx, result.Error)
				return
			}
			// if the application already has a review, replace it with the copied review.
		} else {
			result = h.DB.Model(&model.Review{}).Where("id = ?", existing[0].ID).Updates(&copied)
			if result.Error != nil {
				h.createFailed(ctx, result.Error)
				return
			}
		}
	}
	ctx.Status(http.StatusNoContent)
}

//
// Review REST resource.
type Review struct {
	Resource
	BusinessCriticality uint   `json:"businessCriticality"`
	EffortEstimate      string `json:"effortEstimate"`
	ProposedAction      string `json:"proposedAction"`
	WorkPriority        uint   `json:"workPriority"`
	Comments            string `json:"comments"`
	Application         Ref    `json:"application" binding:"required"`
}

// With updates the resource with the model.
func (r *Review) With(m *model.Review) {
	r.Resource.With(&m.Model)
	r.BusinessCriticality = m.BusinessCriticality
	r.EffortEstimate = m.EffortEstimate
	r.ProposedAction = m.ProposedAction
	r.WorkPriority = m.WorkPriority
	r.Comments = m.Comments
	r.Application.ID = m.Application.ID
	r.Application.Name = m.Application.Name
}

//
// Model builds a model.
func (r *Review) Model() (m *model.Review) {
	m = &model.Review{
		BusinessCriticality: r.BusinessCriticality,
		EffortEstimate:      r.EffortEstimate,
		ProposedAction:      r.ProposedAction,
		WorkPriority:        r.WorkPriority,
		Comments:            r.Comments,
		ApplicationID:       r.Application.ID,
	}
	m.ID = r.ID
	return
}

//
// CopyRequest REST resource.
type CopyRequest struct {
	SourceReview       uint   `json:"sourceReview" binding:"required"`
	TargetApplications []uint `json:"targetApplications" binding:"required"`
}
