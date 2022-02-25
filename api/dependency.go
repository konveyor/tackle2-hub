package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
)

//
// Routes
const (
	DependenciesRoot = "/dependencies"
	DependencyRoot   = DependenciesRoot + "/:" + ID
)

//
// DependencyHandler handles application dependency routes.
type DependencyHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h DependencyHandler) AddRoutes(e *gin.Engine) {
	e.GET(DependenciesRoot, h.List)
	e.GET(DependenciesRoot+"/", h.List)
	e.POST(DependenciesRoot, h.Create)
	e.GET(DependencyRoot, h.Get)
	e.DELETE(DependencyRoot, h.Delete)
}

// Get godoc
// @summary Get a dependency by ID.
// @description Get a dependency by ID.
// @tags get
// @produce json
// @success 200 {object} api.Dependency
// @router /dependencies/{id} [get]
// @param id path string true "Dependency ID"
func (h DependencyHandler) Get(ctx *gin.Context) {
	m := &model.Dependency{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB, "To", "From")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Dependency{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

//
// List godoc
// @summary List all dependencies.
// @description List all dependencies.
// @tags list
// @produce json
// @success 200 {object} []api.Dependency
// @router /dependencies [get]
func (h DependencyHandler) List(ctx *gin.Context) {
	var list []model.Dependency

	db := h.DB
	to := ctx.Query("to.id")
	from := ctx.Query("from.id")
	if to != "" {
		db = db.Where("toid = ?", to)
	} else if from != "" {
		db = db.Where("fromid = ?", from)
	}

	db = h.preLoad(db, "To", "From")
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}

	resources := []Dependency{}
	for i := range list {
		r := Dependency{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a dependency.
// @description Create a dependency.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Dependency
// @router /dependencies [post]
// @param applications_dependency body Dependency true "Dependency data"
func (h DependencyHandler) Create(ctx *gin.Context) {
	r := Dependency{}
	err := ctx.BindJSON(&r)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Create(&m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a dependency.
// @description Delete a dependency.
// @tags delete
// @accept json
// @success 204
// @router /dependencies/{id} [delete]
// @param id path string true "Dependency id"
func (h DependencyHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	result := h.DB.Delete(&model.Dependency{}, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Dependency REST resource.
type Dependency struct {
	Resource
	To   Ref `json:"to"`
	From Ref `json:"from"`
}

//
// With updates the resource using the model.
func (r *Dependency) With(m *model.Dependency) {
	r.Resource.With(&m.Model)
	r.To.ID = m.ToID
	r.To.Name = m.To.Name
	r.From.ID = m.FromID
	r.From.Name = m.From.Name
}

// Model builds a model.Dependency.
func (r *Dependency) Model() (m *model.Dependency) {
	m = &model.Dependency{
		ToID:   r.To.ID,
		FromID: r.From.ID,
	}
	return
}
