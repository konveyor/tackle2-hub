package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Routes
const (
	AnalysisProfilesRoot  = "/analysis/profiles"
	AnalysisProfileRoot   = AnalysisProfilesRoot + "/:id"
	AnalysisProfileBundle = AnalysisProfileRoot + "/:bundle"
)

// AnalysisProfileHandler handles application Profile resource routes.
type AnalysisProfileHandler struct {
	BaseHandler
}

func (h AnalysisProfileHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("Profiles"))
	routeGroup.GET(AnalysisProfileRoot, h.Get)
	routeGroup.GET(AnalysisProfilesRoot, h.List)
	routeGroup.GET(AnalysisProfilesRoot+"/", h.List)
	routeGroup.POST(AnalysisProfilesRoot, h.Create)
	routeGroup.PUT(AnalysisProfileRoot, h.Update)
	routeGroup.DELETE(AnalysisProfileRoot, h.Delete)
}

// Get godoc
// @summary Get a Profile by ID.
// @description Get a Profile by ID.
// @tags Profiles
// @produce json
// @success 200 {object} AnalysisProfile
// @router /Profiles/{id} [get]
// @param id path int true "Profile ID"
func (h AnalysisProfileHandler) Get(ctx *gin.Context) {
	r := AnalysisProfile{}
	id := h.pk(ctx)
	m := &model.AnalysisProfile{}
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all Profiles.
// @description List all Profiles.
// @tags Profiles
// @produce json
// @success 200 {object} []AnalysisProfile
// @router /Profiles [get]
func (h AnalysisProfileHandler) List(ctx *gin.Context) {
	resources := []AnalysisProfile{}
	var list []model.AnalysisProfile
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		r := AnalysisProfile{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a Profile.
// @description Create a Profile.
// @tags Profiles
// @accept json
// @produce json
// @success 201 {object} Profile
// @router /Profiles [post]
// @param Profile body AnalysisProfile true "Profile data"
func (h AnalysisProfileHandler) Create(ctx *gin.Context) {
	r := &AnalysisProfile{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	db = db.Omit(clause.Associations)
	err = db.Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Targets").Replace(m.Targets)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a Profile.
// @description Delete a Profile.
// @tags Profiles
// @success 204
// @router /Profiles/{id} [delete]
// @param id path int true "Profile ID"
func (h AnalysisProfileHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.AnalysisProfile{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = db.Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a Profile.
// @description Update a Profile.
// @tags Profiles
// @accept json
// @success 204
// @router /Profiles/{id} [put]
// @param id path int true "Profile ID"
// @param Profile body AnalysisProfile true "Profile data"
func (h AnalysisProfileHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &AnalysisProfile{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

type InExList = model.InExList

// AnalysisProfile REST resource.
type AnalysisProfile struct {
	Resource `yaml:",inline"`
	Name     string `json:"name"`
	Mode     struct {
		WithDeps bool `json:"withDeps" yaml:"withDeps"`
	} `json:"mode"`
	Scope struct {
		WithKnownLibs bool     `json:"withKnownLibs" yaml:"withKnownLibs"`
		Packages      InExList `json:"packages"`
	} `json:"scope"`
	Rules struct {
		Targets    []Ref      `json:"targets"`
		Labels     InExList   `json:"labels"`
		Files      []Ref      `json:"files"`
		Repository Repository `json:"repository"`
	}
}

// With updates the resource with the model.
func (r *AnalysisProfile) With(m *model.AnalysisProfile) {
	r.Resource.With(&m.Model)
	r.Mode.WithDeps = m.WithDeps
	r.Scope.WithKnownLibs = m.WithKnownLibs
	r.Scope.Packages = m.Packages
	r.Rules.Labels = m.Labels
	r.Rules.Repository = Repository(m.Repository)
	r.Rules.Targets = make([]Ref, len(m.Targets))
	for i, t := range m.Targets {
		r.Rules.Targets[i] =
			Ref{
				ID:   t.ID,
				Name: t.Name,
			}
	}
	r.Rules.Files = make([]Ref, len(m.Files))
	for i, f := range m.Files {
		r.Rules.Files[i] = Ref(f)
	}
}

// Model builds a model.
func (r *AnalysisProfile) Model() (m *model.AnalysisProfile) {
	m = &model.AnalysisProfile{}
	m.WithDeps = r.Mode.WithDeps
	m.WithKnownLibs = r.Scope.WithKnownLibs
	m.Packages = r.Scope.Packages
	m.Labels = r.Rules.Labels
	m.Repository = model.Repository(r.Rules.Repository)
	m.Targets = make([]model.Target, len(r.Rules.Targets))
	for i, t := range r.Rules.Targets {
		m.Targets[i] =
			model.Target{
				Model: model.Model{
					ID: t.ID,
				},
				Name: t.Name,
			}
	}
	m.Files = make([]model.Ref, len(r.Rules.Files))
	for i, f := range r.Rules.Files {
		m.Files[i] = model.Ref(f)
	}
	return
}
