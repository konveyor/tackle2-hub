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
	JobFunctionsRoot = "/jobfunctions"
	JobFunctionRoot  = JobFunctionsRoot + "/:" + ID
)

//
// JobFunctionHandler handles job-function routes.
type JobFunctionHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h JobFunctionHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("jobfunctions"))
	routeGroup.GET(JobFunctionsRoot, h.List)
	routeGroup.GET(JobFunctionsRoot+"/", h.List)
	routeGroup.POST(JobFunctionsRoot, h.Create)
	routeGroup.GET(JobFunctionRoot, h.Get)
	routeGroup.PUT(JobFunctionRoot, h.Update)
	routeGroup.DELETE(JobFunctionRoot, h.Delete)
}

// Get godoc
// @summary Get a job function by ID.
// @description Get a job function by ID.
// @tags jobfunctions
// @produce json
// @success 200 {object} api.JobFunction
// @router /jobfunctions/{id} [get]
// @param id path string true "Job Function ID"
func (h JobFunctionHandler) Get(ctx *gin.Context) {
	m := &model.JobFunction{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := JobFunction{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all job functions.
// @description List all job functions.
// @tags jobfunctions
// @produce json
// @success 200 {object} []api.JobFunction
// @router /jobfunctions [get]
func (h JobFunctionHandler) List(ctx *gin.Context) {
	var list []model.JobFunction
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []JobFunction{}
	for i := range list {
		r := JobFunction{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a job function.
// @description Create a job function.
// @tags jobfunctions
// @accept json
// @produce json
// @success 200 {object} api.JobFunction
// @router /jobfunctions [post]
// @param job_function body api.JobFunction true "Job Function data"
func (h JobFunctionHandler) Create(ctx *gin.Context) {
	r := &JobFunction{}
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
// @summary Delete a job function.
// @description Delete a job function.
// @tags jobfunctions
// @success 204
// @router /jobfunctions/{id} [delete]
// @param id path string true "Job Function ID"
func (h JobFunctionHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.JobFunction{}
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
// @summary Update a job function.
// @description Update a job function.
// @tags jobfunctions
// @accept json
// @success 204
// @router /jobfunctions/{id} [put]
// @param id path string true "Job Function ID"
// @param job_function body api.JobFunction true "Job Function data"
func (h JobFunctionHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &JobFunction{}
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
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// JobFunction REST resource.
type JobFunction struct {
	Resource     `yaml:",inline"`
	Name         string `json:"name" binding:"required"`
	Stakeholders []Ref  `json:"stakeholders"`
}

//
// With updates the resource with the model.
func (r *JobFunction) With(m *model.JobFunction) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
}

//
// Model builds a model.
func (r *JobFunction) Model() (m *model.JobFunction) {
	m = &model.JobFunction{
		Name: r.Name,
	}
	m.ID = r.ID

	return
}
