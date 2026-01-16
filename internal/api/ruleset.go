package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/internal/api/filter"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RuleSetHandler handles ruleset resource routes.
type RuleSetHandler struct {
	BaseHandler
}

func (h RuleSetHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("rulesets"), Transaction)
	routeGroup.GET(api.RuleSetsRoute, h.List)
	routeGroup.GET(api.RuleSetsRoute+"/", h.List)
	routeGroup.POST(api.RuleSetsRoute, h.Create)
	routeGroup.GET(api.RuleSetRoute, h.Get)
	routeGroup.PUT(api.RuleSetRoute, h.Update)
	routeGroup.DELETE(api.RuleSetRoute, h.Delete)
}

// Get godoc
// @summary Get a RuleSet by ID.
// @description Get a RuleSet by ID.
// @tags rulesets
// @produce json
// @success 200 {object} RuleSet
// @router /rulesets/{id} [get]
// @param id path int true "RuleSet ID"
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
		_ = ctx.Error(err)
		return
	}
	err = h.create(ctx, ruleset)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Respond(ctx, http.StatusCreated, ruleset)
}

// Delete godoc
// @summary Delete a ruleset.
// @description Delete a ruleset.
// @tags rulesets
// @success 204
// @router /rulesets/{id} [delete]
// @param id path int true "RuleSet ID"
func (h RuleSetHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	err := h.delete(ctx, id)
	if err != nil {
		_ = ctx.Error(err)
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
// @param id path int true "RuleSet ID"
// @param ruleBundle body RuleSet true "RuleSet data"
func (h RuleSetHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &RuleSet{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.ID = id
	err = h.update(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
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

// create the ruleset.
func (h *RuleSetHandler) create(ctx *gin.Context, r *RuleSet) (err error) {
	m := r.Model()
	err = h.DB(ctx).Create(m).Error
	if err != nil {
		return
	}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Rules.File")
	err = db.First(m).Error
	if err != nil {
		return
	}
	r.With(m)
	return
}

// update the ruleset.
func (h *RuleSetHandler) update(ctx *gin.Context, r *RuleSet) (err error) {
	m := &model.RuleSet{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err = db.First(m, r.ID).Error
	if err != nil {
		return
	}
	if m.Builtin() {
		err = &Forbidden{Reason: "update on builtin not permitted."}
		return
	}
	//
	// Delete unwanted rules.
	for _, rule := range m.Rules {
		if !r.HasRule(rule.ID) {
			err = h.DB(ctx).Delete(rule).Error
			if err != nil {
				return
			}
		}
	}
	//
	// Update ruleset.
	m = r.Model()
	m.ID = r.ID
	m.UpdateUser = h.CurrentUser(ctx)
	db = h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	err = db.Save(m).Error
	if err != nil {
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
		err = db.Updates(m).Error
		if err != nil {
			return
		}
	}
	return
}

// delete the ruleset.
func (h *RuleSetHandler) delete(ctx *gin.Context, id uint) (err error) {
	ruleset := &model.RuleSet{}
	err = h.DB(ctx).First(ruleset, id).Error
	if err != nil {
		return
	}
	if ruleset.Builtin() {
		err = &Forbidden{Reason: "delete on builtin not permitted."}
		return
	}
	err = h.DB(ctx).Delete(ruleset, id).Error
	if err != nil {
		return
	}
	return
}

// RuleSet REST resource.
type RuleSet = resource.RuleSet

// Rule - REST Resource.
type Rule = resource.Rule
