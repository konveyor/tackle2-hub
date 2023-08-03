package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
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
// @description filters:
// @description - name
// @description - labels
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

	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "name", Kind: qf.STRING},
			{Field: "labels", Kind: qf.STRING, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	result := db.Where("ID IN (?)", h.ruleSetIDs(ctx, filter)).Find(&list)
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
	if ruleset.Builtin() {
		h.Status(ctx, http.StatusForbidden)
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
	m := &model.RuleSet{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if m.Builtin() {
		h.Status(ctx, http.StatusForbidden)
		return
	}
	//
	// Delete unwanted rules.
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
	err = h.DB(ctx).Model(m).Association("DependsOn").Replace(m.DependsOn)
	if err != nil {
		_ = ctx.Error(err)
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

func (h *RuleSetHandler) ruleSetIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.RuleSet{})
	q = q.Select("ID")
	q = f.Where(q, "-Labels")
	filter := f
	if f, found := filter.Field("labels"); found {
		if f.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, f = range f.Expand() {
				f = f.As("json_each.value")
				iq := h.DB(ctx)
				iq = iq.Table("Rule")
				iq = iq.Joins("m ,json_each(Labels)")
				iq = iq.Select("m.RuleSetID")
				qs = append(qs, iq)
			}
			q = q.Where("ID IN (?)", model.Intersect(qs...))
		} else {
			f = f.As("json_each.value")
			iq := h.DB(ctx)
			iq = iq.Table("Rule")
			iq = iq.Joins("m ,json_each(Labels)")
			iq = iq.Select("m.RuleSetID")
			iq = f.Where(iq)
			q = q.Where("ID IN (?)", iq)
		}
	}
	return
}

//
// RuleSet REST resource.
type RuleSet struct {
	Resource
	Kind        string      `json:"kind,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Rules       []Rule      `json:"rules"`
	Repository  *Repository `json:"repository,omitempty"`
	Identity    *Ref        `json:"identity,omitempty"`
	DependsOn   []Ref       `json:"dependsOn"`
}

//
// With updates the resource with the model.
func (r *RuleSet) With(m *model.RuleSet) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Name = m.Name
	r.Description = m.Description
	r.Identity = r.refPtr(m.IdentityID, m.Identity)
	_ = json.Unmarshal(m.Repository, &r.Repository)
	r.Rules = []Rule{}
	for i := range m.Rules {
		rule := Rule{}
		rule.With(&m.Rules[i])
		r.Rules = append(
			r.Rules,
			rule)
	}
	r.DependsOn = []Ref{}
	for i := range m.DependsOn {
		dep := Ref{}
		dep.With(m.DependsOn[i].ID, m.DependsOn[i].Name)
		r.DependsOn = append(r.DependsOn, dep)
	}
}

//
// Model builds a model.
func (r *RuleSet) Model() (m *model.RuleSet) {
	m = &model.RuleSet{
		Kind:        r.Kind,
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	m.IdentityID = r.idPtr(r.Identity)
	m.Rules = []model.Rule{}
	for _, rule := range r.Rules {
		m.Rules = append(m.Rules, *rule.Model())
	}
	if r.Repository != nil {
		m.Repository, _ = json.Marshal(r.Repository)
	}
	for _, ref := range r.DependsOn {
		m.DependsOn = append(
			m.DependsOn,
			model.RuleSet{
				Model: model.Model{
					ID: ref.ID,
				},
			})
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
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	File        *Ref     `json:"file,omitempty"`
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
