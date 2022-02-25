package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
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
	e.GET(JobFunctionsRoot, h.List)
	e.GET(JobFunctionsRoot+"/", h.List)
	e.POST(JobFunctionsRoot, h.Create)
	e.GET(JobFunctionRoot, h.Get)
	e.PUT(JobFunctionRoot, h.Update)
	e.DELETE(JobFunctionRoot, h.Delete)
}

// Get godoc
// @summary Get a job function by ID.
// @description Get a job function by ID.
// @tags get
// @produce json
// @success 200 {object} []api.JobFunction
// @router /jobfunctions/{id} [get]
// @param id path string true "Job Function ID"
func (h JobFunctionHandler) Get(ctx *gin.Context) {
	m := &model.JobFunction{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB, "Stakeholders")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := JobFunction{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all job functions.
// @description List all job functions.
// @tags get
// @produce json
// @success 200 {object} []api.JobFunction
// @router /jobfunctions [get]
func (h JobFunctionHandler) List(ctx *gin.Context) {
	var list []model.JobFunction
	db := h.preLoad(h.DB, "Stakeholders")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []JobFunction{}
	for i := range list {
		r := JobFunction{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a job function.
// @description Create a job function.
// @tags create
// @accept json
// @produce json
// @success 200 {object} api.JobFunction
// @router /jobfunctions [post]
// @param job_function body api.JobFunction true "Job Function data"
func (h JobFunctionHandler) Create(ctx *gin.Context) {
	r := &JobFunction{}
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
// @summary Delete a job function.
// @description Delete a job function.
// @tags delete
// @success 204
// @router /jobfunctions/{id} [delete]
// @param id path string true "Job Function ID"
func (h JobFunctionHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	result := h.DB.Delete(&model.JobFunction{}, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a job function.
// @description Update a job function.
// @tags update
// @accept json
// @success 204
// @router /jobfunctions/{id} [put]
// @param id path string true "Job Function ID"
// @param job_function body api.JobFunction true "Job Function data"
func (h JobFunctionHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	r := &JobFunction{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Model(&JobFunction{}).Where("id = ?", id).Omit("id").Updates(m)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// JobFunction REST resource.
type JobFunction struct {
	Resource
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
