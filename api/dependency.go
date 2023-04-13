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
	routeGroup := e.Group("/")
	routeGroup.Use(Required("dependencies"))
	routeGroup.GET(DependenciesRoot, h.List)
	routeGroup.GET(DependenciesRoot+"/", h.List)
	routeGroup.POST(DependenciesRoot, h.Create)
	routeGroup.GET(DependencyRoot, h.Get)
	routeGroup.DELETE(DependencyRoot, h.Delete)
}

// Get godoc
// @summary Get a dependency by ID.
// @description Get a dependency by ID.
// @tags dependencies
// @produce json
// @success 200 {object} api.Dependency
// @router /dependencies/{id} [get]
// @param id path string true "Dependency ID"
func (h DependencyHandler) Get(ctx *gin.Context) {
	m := &model.Dependency{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Dependency{}
	r.With(m)

	h.Render(ctx, http.StatusOK, r)
}

//
// List godoc
// @summary List all dependencies.
// @description List all dependencies.
// @tags dependencies
// @produce json
// @success 200 {object} []api.Dependency
// @router /dependencies [get]
func (h DependencyHandler) List(ctx *gin.Context) {
	var list []model.Dependency

	db := h.Paginated(ctx)
	to := ctx.Query("to.id")
	from := ctx.Query("from.id")
	if to != "" {
		db = db.Where("toid = ?", to)
	} else if from != "" {
		db = db.Where("fromid = ?", from)
	}

	db = h.preLoad(db, clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	resources := []Dependency{}
	for i := range list {
		r := Dependency{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a dependency.
// @description Create a dependency.
// @tags dependencies
// @accept json
// @produce json
// @success 201 {object} api.Dependency
// @router /dependencies [post]
// @param applications_dependency body Dependency true "Dependency data"
func (h DependencyHandler) Create(ctx *gin.Context) {
	r := Dependency{}
	err := h.Bind(ctx, &r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	err = m.Create(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Render(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a dependency.
// @description Delete a dependency.
// @tags dependencies
// @accept json
// @success 204
// @router /dependencies/{id} [delete]
// @param id path string true "Dependency id"
func (h DependencyHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Dependency{}
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
	r.To = r.ref(m.ToID, m.To)
	r.From = r.ref(m.FromID, m.From)
}

// Model builds a model.Dependency.
func (r *Dependency) Model() (m *model.Dependency) {
	m = &model.Dependency{
		ToID:   r.To.ID,
		FromID: r.From.ID,
	}
	return
}
