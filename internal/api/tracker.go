package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/tracker"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// Params
const (
	Connected = "connected"
)

// TrackerHandler handles ticket tracker routes.
type TrackerHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h TrackerHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("trackers"))
	routeGroup.GET(api.TrackersRoute, h.List)
	routeGroup.GET(api.TrackersRoute+"/", h.List)
	routeGroup.POST(api.TrackersRoute, h.Create)
	routeGroup.GET(api.TrackerRoute, h.Get)
	routeGroup.PUT(api.TrackerRoute, h.Update)
	routeGroup.DELETE(api.TrackerRoute, h.Delete)
	routeGroup.GET(api.TrackerProjectsRoute, h.ProjectList)
	routeGroup.GET(api.TrackerProjectRoute, h.ProjectGet)
	routeGroup.GET(api.TrackerProjectIssueTypesRoute, h.ProjectIssueTypeList)
}

// Get godoc
// @summary Get a tracker by ID.
// @description Get a tracker by ID.
// @tags trackers
// @produce json
// @success 200 {object} api.Tracker
// @router /trackers/{id} [get]
// @param id path int true "Tracker ID"
func (h TrackerHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := Tracker{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all trackers.
// @description List all trackers.
// @tags trackers
// @produce json
// @success 200 {object} []api.Tracker
// @router /trackers [get]
func (h TrackerHandler) List(ctx *gin.Context) {
	var list []model.Tracker
	db := h.preLoad(h.DB(ctx), clause.Associations)
	kind := ctx.Query(Kind)
	if kind != "" {
		db = db.Where(Kind, kind)
	}
	q := ctx.Query(Connected)
	if q != "" {
		connected, err := strconv.ParseBool(q)
		if err != nil {
			h.Status(ctx, http.StatusBadRequest)
			return
		}
		db = db.Where(Connected, connected)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Tracker{}
	for i := range list {
		r := Tracker{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a tracker.
// @description Create a tracker.
// @tags trackers
// @accept json
// @produce json
// @success 201 {object} api.Tracker
// @router /trackers [post]
// @param tracker body api.Tracker true "Tracker data"
func (h TrackerHandler) Create(ctx *gin.Context) {
	r := &Tracker{}
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
// @summary Delete a tracker.
// @description Delete a tracker.
// @tags trackers
// @success 204
// @router /trackers/{id} [delete]
// @param id path int true "Tracker id"
func (h TrackerHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
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
// @summary Update a tracker.
// @description Update a tracker.
// @tags trackers
// @accept json
// @success 204
// @router /trackers/{id} [put]
// @param id path int true "Tracker id"
// @param application body api.Tracker true "Tracker data"
func (h TrackerHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Tracker{}
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

// ProjectList godoc
// @summary List a tracker's projects.
// @description List a tracker's projects.
// @tags trackers
// @produce json
// @success 200 {object} []api.Project
// @router /trackers/{id}/projects [get]
// @param id path int true "Tracker ID"
func (h TrackerHandler) ProjectList(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.Connected {
		_ = ctx.Error(&TrackerError{m.Message})
		return
	}
	conn, err := tracker.NewConnector(m)
	if err != nil {
		_ = ctx.Error(&TrackerError{err.Error()})
		return
	}
	projects, err := conn.Projects()
	if err != nil {
		_ = ctx.Error(&TrackerError{err.Error()})
		return
	}

	resources := []Project{}
	for i := range projects {
		r := Project(projects[i])
		resources = append(resources, r)
	}
	h.Respond(ctx, http.StatusOK, resources)
}

// ProjectGet godoc
// @summary Get a tracker project by ID.
// @description Get a tracker project by ID.
// @tags trackers
// @produce json
// @success 200 {object} api.Project
// @router /trackers/{id}/projects/{id2} [get]
// @param id path int true "Tracker ID"
// @param id2 path string true "Project ID"
func (h TrackerHandler) ProjectGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.Connected {
		_ = ctx.Error(&TrackerError{m.Message})
		return
	}
	conn, err := tracker.NewConnector(m)
	if err != nil {
		_ = ctx.Error(&TrackerError{err.Error()})
		return
	}
	project, err := conn.Project(ctx.Param(ID2))
	if err != nil {
		_ = ctx.Error(&TrackerError{err.Error()})
		return
	}

	r := Project(project)
	h.Respond(ctx, http.StatusOK, r)
}

// ProjectIssueTypeList godoc
// @summary List a tracker project's issue types.
// @description List a tracker project's issue types.
// @tags trackers
// @produce json
// @success 200 {object} []api.IssueType
// @router /trackers/{id}/projects/{id2}/issuetypes [get]
// @param id path int true "Tracker ID"
// @param id2 path string true "Project ID"
func (h TrackerHandler) ProjectIssueTypeList(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Tracker{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.Connected {
		_ = ctx.Error(&TrackerError{m.Message})
		return
	}
	conn, err := tracker.NewConnector(m)
	if err != nil {
		_ = ctx.Error(&TrackerError{err.Error()})
		return
	}
	issueTypes, err := conn.IssueTypes(ctx.Param(ID2))
	if err != nil {
		_ = ctx.Error(&TrackerError{err.Error()})
		return
	}

	resources := []IssueType{}
	for i := range issueTypes {
		r := IssueType(issueTypes[i])
		resources = append(resources, r)
	}
	h.Respond(ctx, http.StatusOK, resources)
}

// Tracker API Resource
type Tracker = resource.Tracker

// Project API Resource
type Project = resource.Project

// IssueType API Resource
type IssueType = resource.IssueType
