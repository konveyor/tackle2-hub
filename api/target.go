package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	TargetsRoot = "/targets"
	TargetRoot  = TargetsRoot + "/:" + ID
)

//
// TargetHandler handles Target resource routes.
type TargetHandler struct {
	BaseHandler
}

func (h TargetHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("targets"), Transaction)
	routeGroup.GET(TargetsRoot, h.List)
	routeGroup.GET(TargetsRoot+"/", h.List)
	routeGroup.POST(TargetsRoot, h.Create)
	routeGroup.GET(TargetRoot, h.Get)
	routeGroup.PUT(TargetRoot, h.Update)
	routeGroup.DELETE(TargetRoot, h.Delete)
}

// Get godoc
// @summary Get a Target by ID.
// @description Get a Target by ID.
// @tags targets
// @produce json
// @success 200 {object} Target
// @router /targets/{id} [get]
// @param id path string true "Target ID"
func (h TargetHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	target := &model.Target{}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"RuleSet.Rules",
		"RuleSet.Rules.File")
	result := db.First(target, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Target{}
	r.With(target)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all targets.
// @description List all targets.
// @tags targets
// @produce json
// @success 200 {object} []Target
// @router /targets [get]
func (h TargetHandler) List(ctx *gin.Context) {
	var list []model.Target
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"RuleSet.Rules",
		"RuleSet.Rules.File")
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Target{}
	for i := range list {
		r := Target{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a target.
// @description Create a target.
// @tags targets
// @accept json
// @produce json
// @success 201 {object} Target
// @router /targets [post]
// @param target body Target true "Target data"
func (h TargetHandler) Create(ctx *gin.Context) {
	target := &Target{}
	err := h.Bind(ctx, target)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := target.Model()
	m.CreateUser = h.CurrentUser(ctx)
	if target.RuleSet != nil {
		ruleset := target.RuleSet
		ruleset.Name = fmt.Sprintf("__Target(%s)", m.Name)
		err := (&RuleSetHandler{}).create(ctx, ruleset)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		m.RuleSetID = &ruleset.ID
	}
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"RuleSet.Rules",
		"RuleSet.Rules.File")
	result = db.First(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	target.With(m)

	h.Respond(ctx, http.StatusCreated, target)
}

// Delete godoc
// @summary Delete a target.
// @description Delete a target.
// @tags targets
// @success 204
// @router /targets/{id} [delete]
// @param id path string true "Target ID"
func (h TargetHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	target := &model.Target{}
	result := h.DB(ctx).First(target, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if target.Builtin() {
		h.Status(ctx, http.StatusForbidden)
		return
	}
	result = h.DB(ctx).Delete(target, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if target.RuleSetID != nil {
		err := (&RuleSetHandler{}).delete(ctx, *target.RuleSetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = nil
			} else {
				return
			}
		}
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a target.
// @description Update a target.
// @tags targets
// @accept json
// @success 204
// @router /targets/{id} [put]
// @param id path string true "Target ID"
// @param target body Target true "Target data"
func (h TargetHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Target{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := &model.Target{}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"RuleSet.Rules",
		"RuleSet.Rules.File")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if m.Builtin() {
		h.Status(ctx, http.StatusForbidden)
		return
	}
	m = r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	if r.RuleSet != nil {
		m.RuleSetID = &r.RuleSet.ID
		err := (&RuleSetHandler{}).update(ctx, r.RuleSet)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}
	db = h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result = db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// Target REST resource.
type Target struct {
	Resource
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Choice      bool     `json:"choice,omitempty"`
	Custom      bool     `json:"custom,omitempty"`
	Labels      []Label  `json:"labels"`
	Image       Ref      `json:"image"`
	RuleSet     *RuleSet `json:"ruleset,omitempty"`
}

type Label struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

//
// With updates the resource with the model.
func (r *Target) With(m *model.Target) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Choice = m.Choice
	r.Custom = !m.Builtin()
	if m.RuleSet != nil {
		r.RuleSet = &RuleSet{}
		r.RuleSet.With(m.RuleSet)
	}
	imgRef := Ref{ID: m.ImageID}
	if m.Image != nil {
		imgRef.Name = m.Image.Name
	}
	r.Image = imgRef
	_ = json.Unmarshal(m.Labels, &r.Labels)
}

//
// Model builds a model.
func (r *Target) Model() (m *model.Target) {
	m = &model.Target{
		Name:        r.Name,
		Description: r.Description,
		Choice:      r.Choice,
	}
	m.ID = r.ID
	m.ImageID = r.Image.ID
	m.Labels, _ = json.Marshal(r.Labels)
	return
}
