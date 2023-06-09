package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	RuleSetsRoot = "/rulesets"
	RuleSetRoot  = RuleSetsRoot + "/:" + ID
)

//
// RuleSetHandler handles ruleset resource routes.
type RuleSetHandler struct {
	BaseHandler
}

func (h RuleSetHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("rulesets"), Transaction)
	routeGroup.GET(RuleSetsRoot, h.List)
	routeGroup.GET(RuleSetsRoot+"/", h.List)
	routeGroup.POST(RuleSetsRoot, h.Create)
	routeGroup.GET(RuleSetRoot, h.Get)
	routeGroup.PUT(RuleSetRoot, h.Update)
	routeGroup.DELETE(RuleSetRoot, h.Delete)
}

// Get godoc
// @summary Get a RuleSet by ID.
// @description Get a RuleSet by ID.
// @tags rulesets
// @produce json
// @success 200 {object} RuleSet
// @router /rulesets/{id} [get]
// @param id path string true "RuleSet ID"
func (h RuleSetHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	ruleset := &model.RuleSet{}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Rules.File")
	result := db.First(ruleset, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := RuleSet{}
	r.With(ruleset)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all bindings.
// @description List all bindings.
// @tags rulesets
// @produce json
// @success 200 {object} []RuleSet
// @router /rulesets [get]
func (h RuleSetHandler) List(ctx *gin.Context) {
	var list []model.RuleSet
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Rules.File")
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []RuleSet{}
	for i := range list {
		r := RuleSet{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a ruleset.
// @description Create a ruleset.
// @tags rulesets
// @accept json
// @produce json
// @success 201 {object} RuleSet
// @router /rulesets [post]
// @param ruleBundle body RuleSet true "RuleSet data"
func (h RuleSetHandler) Create(ctx *gin.Context) {
	ruleset := &RuleSet{}
	err := h.Bind(ctx, ruleset)
	if err != nil {
		return
	}
	m := ruleset.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Rules.File")
	result = db.First(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	ruleset.With(m)

	h.Respond(ctx, http.StatusCreated, ruleset)
}

// Delete godoc
// @summary Delete a ruleset.
// @description Delete a ruleset.
// @tags rulesets
// @success 204
// @router /rulesets/{id} [delete]
// @param id path string true "RuleSet ID"
func (h RuleSetHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	ruleset := &model.RuleSet{}
	result := h.DB(ctx).First(ruleset, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(ruleset, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a ruleset.
// @description Update a ruleset.
// @tags rulesets
// @accept json
// @success 204
// @router /rulesets/{id} [put]
// @param id path string true "RuleSet ID"
// @param ruleBundle body RuleSet true "RuleSet data"
func (h RuleSetHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &RuleSet{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Delete unwanted ruleSets.
	m := &model.RuleSet{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for _, ruleset := range m.Rules {
		if !r.HasRule(ruleset.ID) {
			err := h.DB(ctx).Delete(ruleset).Error
			if err != nil {
				_ = ctx.Error(err)
				return
			}
		}
	}
	//
	// Update ruleset.
	m = r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db = h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result = db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.DB(ctx).Model(m).Association("Rules").Replace(m.Rules)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Update ruleSets.
	for i := range m.Rules {
		m := &m.Rules[i]
		db = h.DB(ctx).Model(m)
		err = db.Updates(h.fields(m)).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// RuleSet REST resource.
type RuleSet struct {
	Resource
	Kind        string      `json:"kind,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       Ref         `json:"image"`
	Rules       []Rule      `json:"rules"`
	Custom      bool        `json:"custom,omitempty"`
	Repository  *Repository `json:"repository,omitempty"`
	Identity    *Ref        `json:"identity,omitempty"`
}

//
// With updates the resource with the model.
func (r *RuleSet) With(m *model.RuleSet) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Description = m.Description
	r.Custom = m.Custom
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	imgRef := Ref{ID: m.ImageID}
	if m.Image != nil {
		imgRef.Name = m.Image.Name
	}
	r.Image = imgRef
	_ = json.Unmarshal(m.Repository, &r.Repository)
	r.Rules = []Rule{}
	for i := range m.Rules {
		rule := Rule{}
		rule.With(&m.Rules[i])
		r.Rules = append(
			r.Rules,
			rule)
	}
}

//
// Model builds a model.
func (r *RuleSet) Model() (m *model.RuleSet) {
	m = &model.RuleSet{
		Kind:        r.Kind,
		Name:        r.Name,
		Description: r.Description,
		Custom:      r.Custom,
	}
	m.ID = r.ID
	m.ImageID = r.Image.ID
	m.IdentityID = r.idPtr(r.Identity)
	m.Rules = []model.Rule{}
	for _, rule := range r.Rules {
		m.Rules = append(m.Rules, *rule.Model())
	}
	if r.Repository != nil {
		m.Repository, _ = json.Marshal(r.Repository)
	}
	return
}

//
// HasRule - determine if the ruleset is referenced.
func (r *RuleSet) HasRule(id uint) (b bool) {
	for _, ruleset := range r.Rules {
		if id == ruleset.ID {
			b = true
			break
		}
	}
	return
}

//
// Rule - REST Resource.
type Rule struct {
	Resource
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Labels      interface{} `json:"labels,omitempty"`
	File        *Ref        `json:"file,omitempty"`
}

//
// With updates the resource with the model.
func (r *Rule) With(m *model.Rule) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	_ = json.Unmarshal(m.Labels, &r.Labels)
	r.File = r.refPtr(m.FileID, m.File)
}

//
// Model builds a model.
func (r *Rule) Model() (m *model.Rule) {
	m = &model.Rule{}
	m.ID = r.ID
	m.Name = r.Name
	if r.Labels != nil {
		m.Labels, _ = json.Marshal(r.Labels)
	}
	m.FileID = r.idPtr(r.File)
	return
}
