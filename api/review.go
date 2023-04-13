package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
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
	routeGroup.Use(Required("reviews"))
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
// @tags reviews
// @produce json
// @success 200 {object} []api.Review
// @router /reviews/{id} [get]
// @param id path string true "Review ID"
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

	h.Render(ctx, http.StatusOK, r)
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
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
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

	h.Render(ctx, http.StatusOK, resources)
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

	h.Render(ctx, http.StatusCreated, review)
}

// Delete godoc
// @summary Delete a review.
// @description Delete a review.
// @tags reviews
// @success 204
// @router /reviews/{id} [delete]
// @param id path string true "Review ID"
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

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a review.
// @description Update a review.
// @tags reviews
// @accept json
// @success 204
// @router /reviews/{id} [put]
// @param id path string true "Review ID"
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
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
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
			ApplicationID:       id,
		}
		existing := []model.Review{}
		result = h.DB(ctx).Find(&existing, "applicationid = ?", id)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
		// if the application doesn't already have a review, create one.
		if len(existing) == 0 {
			result = h.DB(ctx).Create(copied)
			if result.Error != nil {
				_ = ctx.Error(result.Error)
				return
			}
			// if the application already has a review, replace it with the copied review.
		} else {
			result = h.DB(ctx).Model(&existing[0]).Updates(h.fields(copied))
			if result.Error != nil {
				_ = ctx.Error(result.Error)
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
	r.Application = r.ref(m.ApplicationID, m.Application)
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
