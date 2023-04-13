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
	RuleBundlesRoot = "/rulebundles"
	RuleBundleRoot  = RuleBundlesRoot + "/:" + ID
)

//
// RuleBundleHandler handles bundle resource routes.
type RuleBundleHandler struct {
	BaseHandler
}

func (h RuleBundleHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("rulebundles"), Transaction)
	routeGroup.GET(RuleBundlesRoot, h.List)
	routeGroup.GET(RuleBundlesRoot+"/", h.List)
	routeGroup.POST(RuleBundlesRoot, h.Create)
	routeGroup.GET(RuleBundleRoot, h.Get)
	routeGroup.PUT(RuleBundleRoot, h.Update)
	routeGroup.DELETE(RuleBundleRoot, h.Delete)
}

// Get godoc
// @summary Get a RuleBundle by ID.
// @description Get a RuleBundle by ID.
// @tags rulebundles
// @produce json
// @success 200 {object} RuleBundle
// @router /rulebundles/{id} [get]
// @param id path string true "RuleBundle ID"
func (h RuleBundleHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	bundle := &model.RuleBundle{}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"RuleSets.File")
	result := db.First(bundle, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := RuleBundle{}
	r.With(bundle)

	h.Render(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all bindings.
// @description List all bindings.
// @tags rulebundles
// @produce json
// @success 200 {object} []RuleBundle
// @router /rulebundles [get]
func (h RuleBundleHandler) List(ctx *gin.Context) {
	var list []model.RuleBundle
	db := h.preLoad(
		h.Paginated(ctx),
		clause.Associations,
		"RuleSets.File")
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []RuleBundle{}
	for i := range list {
		r := RuleBundle{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a bundle.
// @description Create a bundle.
// @tags rulebundles
// @accept json
// @produce json
// @success 201 {object} RuleBundle
// @router /rulebundles [post]
// @param ruleBundle body RuleBundle true "RuleBundle data"
func (h RuleBundleHandler) Create(ctx *gin.Context) {
	bundle := &RuleBundle{}
	err := h.Bind(ctx, bundle)
	if err != nil {
		return
	}
	m := bundle.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"RuleSets.File")
	result = db.First(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	bundle.With(m)

	h.Render(ctx, http.StatusCreated, bundle)
}

// Delete godoc
// @summary Delete a bundle.
// @description Delete a bundle.
// @tags rulebundles
// @success 204
// @router /rulebundles/{id} [delete]
// @param id path string true "RuleBundle ID"
func (h RuleBundleHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	bundle := &model.RuleBundle{}
	result := h.DB(ctx).First(bundle, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(bundle, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a bundle.
// @description Update a bundle.
// @tags rulebundles
// @accept json
// @success 204
// @router /rulebundles/{id} [put]
// @param id path string true "RuleBundle ID"
// @param ruleBundle body RuleBundle true "RuleBundle data"
func (h RuleBundleHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &RuleBundle{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Delete unwanted ruleSets.
	m := &model.RuleBundle{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for _, ruleset := range m.RuleSets {
		if !r.HasRuleSet(ruleset.ID) {
			err := h.DB(ctx).Delete(ruleset).Error
			if err != nil {
				_ = ctx.Error(err)
				return
			}
		}
	}
	//
	// Update bundle.
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
	err = h.DB(ctx).Model(m).Association("RuleSets").Replace(m.RuleSets)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Update ruleSets.
	for i := range m.RuleSets {
		m := &m.RuleSets[i]
		db = h.DB(ctx).Model(m)
		err = db.Updates(h.fields(m)).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}

	ctx.Status(http.StatusNoContent)
}

//
// RuleBundle REST resource.
type RuleBundle struct {
	Resource
	Kind        string      `json:"kind,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       Ref         `json:"image"`
	RuleSets    []RuleSet   `json:"rulesets"`
	Custom      bool        `json:"custom,omitempty"`
	Repository  *Repository `json:"repository,omitempty"`
	Identity    *Ref        `json:"identity,omitempty"`
}

//
// With updates the resource with the model.
func (r *RuleBundle) With(m *model.RuleBundle) {
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
	r.RuleSets = []RuleSet{}
	for i := range m.RuleSets {
		rule := RuleSet{}
		rule.With(&m.RuleSets[i])
		r.RuleSets = append(
			r.RuleSets,
			rule)
	}
}

//
// Model builds a model.
func (r *RuleBundle) Model() (m *model.RuleBundle) {
	m = &model.RuleBundle{
		Kind:        r.Kind,
		Name:        r.Name,
		Description: r.Description,
		Custom:      r.Custom,
	}
	m.ID = r.ID
	m.ImageID = r.Image.ID
	m.IdentityID = r.idPtr(r.Identity)
	m.RuleSets = []model.RuleSet{}
	for _, rule := range r.RuleSets {
		m.RuleSets = append(m.RuleSets, *rule.Model())
	}
	if r.Repository != nil {
		m.Repository, _ = json.Marshal(r.Repository)
	}
	return
}

//
// HasRuleSet - determine if the ruleset is referenced.
func (r *RuleBundle) HasRuleSet(id uint) (b bool) {
	for _, ruleset := range r.RuleSets {
		if id == ruleset.ID {
			b = true
			break
		}
	}
	return
}

//
// RuleSet - REST Resource.
type RuleSet struct {
	Resource
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
	File        *Ref        `json:"file,omitempty"`
}

//
// With updates the resource with the model.
func (r *RuleSet) With(m *model.RuleSet) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	_ = json.Unmarshal(m.Metadata, &r.Metadata)
	r.File = r.refPtr(m.FileID, m.File)
}

//
// Model builds a model.
func (r *RuleSet) Model() (m *model.RuleSet) {
	m = &model.RuleSet{}
	m.ID = r.ID
	m.Name = r.Name
	if r.Metadata != nil {
		m.Metadata, _ = json.Marshal(r.Metadata)
	}
	m.FileID = r.idPtr(r.File)
	return
}
