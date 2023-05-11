package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"net/http"
	"strings"
)

//
// Routes
const (
	AnalysesRoot               = "/analyses"
	AnalysisRoot               = AnalysesRoot + "/:" + ID
	AnalysesDepsRoot           = AnalysesRoot + "/dependencies"
	AnalysesIssuesRoot         = AnalysesRoot + "/issues"
	AnalysesCompositeRoot      = AnalysesRoot + "/composite"
	AnalysisCompositeDepRoot   = AnalysesCompositeRoot + "/dependencies"
	AnalysisCompositeIssueRoot = AnalysesCompositeRoot + "/issues"

	AppAnalysesRoot       = ApplicationRoot + "/analyses"
	AppAnalysisRoot       = ApplicationRoot + "/analysis"
	AppAnalysisDepsRoot   = AppAnalysisRoot + "/dependencies"
	AppAnalysisIssuesRoot = AppAnalysisRoot + "/issues"
)

//
// AnalysisHandler handles analysis resource routes.
type AnalysisHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h AnalysisHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("application"))
	//
	routeGroup.GET(AnalysisRoot, h.Get)
	routeGroup.DELETE(AnalysisRoot, h.Delete)
	routeGroup.GET(AnalysesDepsRoot, h.Deps)
	routeGroup.GET(AnalysesIssuesRoot, h.Issues)
	routeGroup.GET(AnalysisCompositeIssueRoot, h.IssueComposites)
	routeGroup.GET(AnalysisCompositeDepRoot, h.DepComposites)
	//
	routeGroup.POST(AppAnalysesRoot, h.AppCreate)
	routeGroup.GET(AppAnalysesRoot, h.AppList)
	routeGroup.GET(AppAnalysisRoot, h.AppLatest)
	routeGroup.GET(AppAnalysisDepsRoot, h.AppDeps)
	routeGroup.GET(AppAnalysisIssuesRoot, h.AppIssues)
}

// Get godoc
// @summary Get an analysis (report) by ID.
// @description Get an analysis (report) by ID.
// @tags analyses
// @produce json
// @success 200 {object} api.Analysis
// @router /analyses/{id} [get]
// @param id path string true "Analysis ID"
func (h AnalysisHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Analysis{}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Issues.Incidents")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Analysis{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// AppLatest godoc
// @summary Get the latest analysis.
// @description Get the latest analysis for an application.
// @tags analyses
// @produce json
// @success 200 {object} api.Analysis
// @router /applications/{id}/analysis [get]
// @param id path string true "Application ID"
func (h AnalysisHandler) AppLatest(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Analysis{}
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Issues.Incidents")
	db = db.Where("ApplicationID = ?", id)
	result := db.Last(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Analysis{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// AppList godoc
// @summary List analyses.
// @description List analyses for an application.
// @description Resources do not include relations.
// @tags analyses
// @produce json
// @success 200 {object} []api.Analysis
// @router /analyses [get]
func (h AnalysisHandler) AppList(ctx *gin.Context) {
	resources := []Analysis{}
	// Build query.
	id := h.pk(ctx)
	db := h.Paginated(ctx)
	db = db.Where("ApplicationID = ?", id)
	count := int64(0)
	// Count.
	result := db.Model(&model.Analysis{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err := h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	var list []model.Analysis
	result = db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for i := range list {
		r := Analysis{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AppCreate godoc
// @summary Create an analysis.
// @description Create an analysis.
// @tags analyses
// @accept json
// @produce json
// @success 201 {object} api.Analysis
// @router /application/{id}/analyses [post]
// @param task body api.Analysis true "Analysis data"
func (h AnalysisHandler) AppCreate(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Analysis{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ApplicationID = id
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx)
	db.Logger = db.Logger.LogMode(logger.Error)
	result := db.Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an analysis by ID.
// @description Delete an analysis by ID.
// @tags analyses
// @success 204
// @router /analyses/{id} [delete]
// @param id path string true "Analysis ID"
func (h AnalysisHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &model.Analysis{}
	result := h.DB(ctx).First(r, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(r, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// AppDeps godoc
// @summary List application dependencies.
// @description List application dependencies.
// @description filters:
// @description - id
// @description - name
// @description - version
// @description - type
// @description - sha
// @description - indirect
// @tags dependencies
// @produce json
// @success 200 {object} []api.AnalysisDependency
// @router /application/{id}/analysis/dependencies [get]
// @param id path string true "Application ID"
func (h AnalysisHandler) AppDeps(ctx *gin.Context) {
	resources := []AnalysisDependency{}
	// Latest.
	id := h.pk(ctx)
	analysis := &model.Analysis{}
	db := h.DB(ctx)
	db = db.Where("ApplicationID = ?", id)
	result := db.Last(analysis)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "name", Kind: qf.STRING},
			{Field: "version", Kind: qf.STRING},
			{Field: "type", Kind: qf.STRING},
			{Field: "indirect", Kind: qf.STRING},
			{Field: "sha", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx)
	db = db.Where("AnalysisID = ?", analysis.ID)
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result = db.Model(&model.AnalysisDependency{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	list := []model.AnalysisDependency{}
	db = h.paginated(ctx, db)
	result = db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for i := range list {
		r := AnalysisDependency{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AppIssues godoc
// @summary List application issues.
// @description List application issues.
// @description filters:
// @description - id
// @description - ruleid
// @description - name
// @description - category
// @description - effort
// @tags issues
// @produce json
// @success 200 {object} []api.AnalysisIssue
// @router /application/{id}/analysis/issues [get]
// @param id path string true "Application ID"
func (h AnalysisHandler) AppIssues(ctx *gin.Context) {
	resources := []AnalysisIssue{}
	// Latest.
	id := h.pk(ctx)
	analysis := &model.Analysis{}
	db := h.DB(ctx).Where("ApplicationID = ?", id)
	result := db.Last(analysis)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "ruleid", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx)
	db = db.Where("AnalysisID = ?", analysis.ID)
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result = db.Model(&model.AnalysisIssue{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	list := []model.AnalysisIssue{}
	db = h.paginated(ctx, db)
	result = db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for i := range list {
		r := AnalysisIssue{}
		r.With(&list[i])
		r.Application = id
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Issues godoc
// @summary List all issues.
// @description List all issues.
// @description filters:
// @description - id
// @description - ruleid
// @description - name
// @description - category
// @description - effort
// @description - labels
// @description - application.(id|name)
// @description - tag.id
// @tags issues
// @produce json
// @success 200 {object} []api.AnalysisIssue
// @router /analyses/issues [get]
func (h AnalysisHandler) Issues(ctx *gin.Context) {
	resources := []AnalysisIssue{}
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "ruleid", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "labels", Kind: qf.STRING},
			{Field: "affected", Kind: qf.LITERAL},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, Relation: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Table("AnalysisIssue i,")
	db = db.Joins("Analysis a")
	db = db.Where("a.ID = i.AnalysisID")
	db = db.Where("a.ID IN (?)", h.analysisIDs(ctx, &filter))
	n, q := h.issueIDs(ctx, &filter)
	if n > 0 {
		db = db.Where("i.ID IN (?)", q)
	}
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result := db.Model(&model.AnalysisIssue{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Find.
	type M struct {
		model.AnalysisIssue
		ApplicationID uint
	}
	db = h.paginated(ctx, db)
	db = db.Select(
		"i.*",
		"a.ApplicationID")
	db = db.Preload(clause.Associations)
	var list []M
	result = db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for i := range list {
		m := &list[i]
		r := AnalysisIssue{}
		r.With(&m.AnalysisIssue)
		r.Application = m.ApplicationID
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// IssueComposites godoc
// @summary List issue composites.
// @description List issue composites.
// @description filters:
// @description - category
// @description - effort
// @description - labels
// @description - application.(id|name)
// @description - tag.id
// @tags issuecomposites
// @produce json
// @success 200 {object} []api.IssueComposite
// @router /analyses/issues [get]
func (h AnalysisHandler) IssueComposites(ctx *gin.Context) {
	resources := []*IssueComposite{}
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "affected", Kind: qf.LITERAL},
			{Field: "labels", Kind: qf.STRING, Relation: true},
			{Field: "application.id", Kind: qf.STRING},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, Relation: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Select(
		"i.RuleID",
		"i.Name",
		"i.Description",
		"i.Category",
		"i.Effort",
		"i.Labels",
		"COUNT(distinct a.ID) Affected")
	db = db.Table("AnalysisIssue i,")
	db = db.Joins("Analysis a")
	db = db.Where("a.ID = i.AnalysisID")
	db = db.Where("a.ID in (?)", h.analysisIDs(ctx, &filter))
	n, q := h.issueIDs(ctx, &filter)
	if n > 0 {
		db = db.Where("i.ID IN (?)", q)
	}
	db = db.Group("i.RuleID")
	db = db.Order("i.RuleID")
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result := db.Model(&model.AnalysisIssue{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	affected := make(map[string]int)
	// Find.
	type M struct {
		model.AnalysisIssue
		Affected int
	}
	var list []M
	result = db.Scan(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for i := range list {
		r := &list[i]
		affected[r.RuleID] = r.Affected
	}

	collated := make(map[string]*IssueComposite)
	for i := range list {
		m := list[i]
		r, found := collated[m.RuleID]
		if !found {
			r = &IssueComposite{
				Affected:    affected[m.RuleID],
				Description: m.Description,
				Category:    m.Category,
				RuleID:      m.RuleID,
				Name:        m.Name,
			}
			collated[m.RuleID] = r
			resources = append(resources, r)
			if m.Labels != nil {
				_ = json.Unmarshal(m.Labels, &r.Labels)
			}
		}
		r.Effort += m.Effort
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Deps godoc
// @summary List dependencies.
// @description List dependencies.
// @description filters:
// @description - name
// @description - version
// @description - type
// @description - sha
// @description - indirect
// @description - application.(id|name)
// @description - tag.id
// @tags dependencies
// @produce json
// @success 200 {object} []api.AnalysisDependency
// @router /analyses/dependencies [get]
func (h AnalysisHandler) Deps(ctx *gin.Context) {
	resources := []AnalysisDependency{}
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "name", Kind: qf.STRING},
			{Field: "version", Kind: qf.STRING},
			{Field: "type", Kind: qf.STRING},
			{Field: "indirect", Kind: qf.STRING},
			{Field: "sha", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, Relation: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Where("AnalysisID IN (?)", h.analysisIDs(ctx, &filter))
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result := db.Model(&model.AnalysisDependency{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	db = h.paginated(ctx, db)
	list := []model.AnalysisDependency{}
	result = db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	for i := range list {
		r := AnalysisDependency{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// DepComposites godoc
// @summary List dependency composites.
// @description List dependency composites.
// @description filters:
// @description - name
// @description - version
// @description - type
// @description - sha
// @description - indirect
// @description - application.(id|name)
// @description - tag.id
// @tags dependencies
// @produce json
// @success 200 {object} []api.AnalysisDependency
// @router /analyses/dependencies [get]
func (h AnalysisHandler) DepComposites(ctx *gin.Context) {
	resources := []DepComposite{}
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "name", Kind: qf.STRING},
			{Field: "version", Kind: qf.STRING},
			{Field: "type", Kind: qf.STRING},
			{Field: "indirect", Kind: qf.STRING},
			{Field: "sha", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, Relation: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Where("AnalysisID IN (?)", h.analysisIDs(ctx, &filter))
	db = filter.Where(db)
	db = db.Select(
		"Name",
		"Version",
		"Type",
		"SHA",
		"COUNT(distinct AnalysisID) Affected")
	db = db.Group(
		strings.Join(
			[]string{
				"Name",
				"Version",
				"Type",
				"SHA",
			},
			","))
	// Count.
	count := int64(0)
	result := db.Model(&model.AnalysisDependency{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Respond(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	db = h.paginated(ctx, db)
	result = db.Find(&resources)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Respond(ctx, http.StatusOK, resources)
}

//
// appIDs provides application IDs.
// filter:
// - application.(id|name)
// - tag.id
func (h *AnalysisHandler) appIDs(ctx *gin.Context, f *qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Application{})
	q = q.Select("ID")
	appFilter := f.Resource("application")
	q = appFilter.Where(q)
	tagFilter := f.Resource("tag")
	if field, found := tagFilter.Field("id"); found {
		if field.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, v := range field.Value.ByKind(qf.LITERAL, qf.STRING) {
				q := h.DB(ctx)
				q = q.Model(&model.ApplicationTag{})
				q = q.Select("applicationID ID")
				q = q.Where("TagID = ?", qf.AsValue(v))
				qs = append(qs, q)
			}
			tq := model.Intersect(qs...)
			q = q.Where("ID IN (?)", tq)
		} else {
			field = field.As("TagID")
			tq := h.DB(ctx)
			tq = tq.Model(&model.ApplicationTag{})
			tq = tq.Select("ApplicationID ID")
			tq = tq.Where(field.SQL())
			q = q.Where("ID IN (?)", tq)
		}
	}
	return
}

//
// analysisIDs provides analysis IDs.
func (h *AnalysisHandler) analysisIDs(ctx *gin.Context, f *qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Analysis{})
	q = q.Select("MAX(ID)")
	q = q.Where("ApplicationID IN (?)", h.appIDs(ctx, f))
	q = q.Group("ApplicationID")
	return
}

//
// issueIDs returns filtered issues IDs query.
// filter:
//   - labels
func (h *AnalysisHandler) issueIDs(ctx *gin.Context, f *qf.Filter) (n int, q *gorm.DB) {
	filter := f
	if f, found := filter.Field("labels"); found {
		n = len(f.Value)
		_ = filter.Delete("labels")
		if f.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, v := range f.Value.ByKind(qf.LITERAL, qf.STRING) {
				q := h.DB(ctx)
				q = q.Model(&model.AnalysisIssue{})
				q = q.Joins("i ,json_each(Labels)")
				q = q.Select("i.ID")
				q = q.Where("json_each.value = ?", qf.AsValue(v))
				qs = append(qs, q)
			}
			q = model.Intersect(qs...)
		} else {
			f = f.As("json_each.value")
			q = h.DB(ctx)
			q = q.Model(&model.AnalysisIssue{})
			q = q.Joins("i ,json_each(Labels)")
			q = q.Select("i.ID")
			q = q.Where(f.SQL())
		}
	}
	return
}

//
// Analysis (Analysis) REST resource.
type Analysis struct {
	Resource     `yaml:",inline"`
	Issues       []AnalysisIssue      `json:"issues"`
	Dependencies []AnalysisDependency `json:"dependencies"`
}

//
// With updates the resource with the model.
func (r *Analysis) With(m *model.Analysis) {
	r.Resource.With(&m.Model)
	r.Issues = []AnalysisIssue{}
	for i := range m.Issues {
		n := AnalysisIssue{}
		n.With(&m.Issues[i])
		r.Issues = append(
			r.Issues,
			n)
	}
	r.Dependencies = []AnalysisDependency{}
	for i := range m.Dependencies {
		n := AnalysisDependency{}
		n.With(&m.Dependencies[i])
		r.Dependencies = append(
			r.Dependencies,
			n)
	}
}

//
// Model builds a model.
func (r *Analysis) Model() (m *model.Analysis) {
	m = &model.Analysis{}
	m.Issues = []model.AnalysisIssue{}
	for i := range r.Issues {
		n := r.Issues[i].Model()
		m.Issues = append(
			m.Issues,
			*n)
	}
	m.Dependencies = []model.AnalysisDependency{}
	for i := range r.Dependencies {
		n := r.Dependencies[i].Model()
		m.Dependencies = append(
			m.Dependencies,
			*n)
	}
	return
}

//
// AnalysisIssue REST resource.
type AnalysisIssue struct {
	Resource    `yaml:",inline"`
	RuleID      string             `json:"ruleId" binding:"-"`
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description,omitempty" yaml:",omitempty"`
	Category    string             `json:"category" binding:"required"`
	Effort      int                `json:"effort,omitempty" yaml:",omitempty"`
	Incidents   []AnalysisIncident `json:"incidents,omitempty" yaml:",omitempty"`
	Links       []AnalysisLink     `json:"links,omitempty" yaml:",omitempty"`
	Facts       FactMap            `json:"facts,omitempty" yaml:",omitempty"`
	Labels      []string           `json:"labels"`
	Application uint               `json:"application" binding:"-"`
}

//
// With updates the resource with the model.
func (r *AnalysisIssue) With(m *model.AnalysisIssue) {
	r.Resource.With(&m.Model)
	r.RuleID = m.RuleID
	r.Name = m.Name
	r.Description = m.Description
	r.Category = m.Category
	r.Incidents = []AnalysisIncident{}
	for i := range m.Incidents {
		n := AnalysisIncident{}
		n.With(&m.Incidents[i])
		r.Incidents = append(
			r.Incidents,
			n)
	}
	if m.Links != nil {
		_ = json.Unmarshal(m.Links, &r.Links)
	}
	if m.Facts != nil {
		_ = json.Unmarshal(m.Facts, &r.Facts)
	}
	if m.Labels != nil {
		_ = json.Unmarshal(m.Labels, &r.Labels)
	}
	r.Effort = m.Effort
}

//
// Model builds a model.
func (r *AnalysisIssue) Model() (m *model.AnalysisIssue) {
	m = &model.AnalysisIssue{}
	m.RuleID = r.RuleID
	m.Name = r.Name
	m.Description = r.Description
	m.Category = r.Category
	m.Incidents = []model.AnalysisIncident{}
	for i := range r.Incidents {
		n := r.Incidents[i].Model()
		m.Incidents = append(
			m.Incidents,
			*n)
	}
	m.Links, _ = json.Marshal(r.Links)
	m.Facts, _ = json.Marshal(r.Facts)
	m.Labels, _ = json.Marshal(r.Labels)
	m.Effort = r.Effort
	return
}

//
// AnalysisDependency REST resource.
type AnalysisDependency struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Version  string `json:"version,omitempty" yaml:",omitempty"`
	Type     string `json:"type,omitempty" yaml:",omitempty"`
	Indirect bool   `json:"indirect,omitempty" yaml:",omitempty"`
	SHA      string `json:"sha,omitempty" yaml:",omitempty"`
}

//
// With updates the resource with the model.
func (r *AnalysisDependency) With(m *model.AnalysisDependency) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Version = m.Version
	r.Type = m.Type
	r.Indirect = m.Indirect
	r.SHA = m.SHA
}

//
// Model builds a model.
func (r *AnalysisDependency) Model() (m *model.AnalysisDependency) {
	m = &model.AnalysisDependency{}
	m.Name = r.Name
	m.Version = r.Version
	m.Type = r.Type
	m.Indirect = r.Indirect
	m.SHA = r.SHA
	return
}

//
// AnalysisIncident REST resource.
type AnalysisIncident struct {
	Resource `yaml:",inline"`
	URI      string  `json:"uri"`
	Message  string  `json:"message"`
	Facts    FactMap `json:"facts"`
}

//
// With updates the resource with the model.
func (r *AnalysisIncident) With(m *model.AnalysisIncident) {
	r.Resource.With(&m.Model)
	r.URI = m.URI
	r.Message = m.Message
	if m.Facts != nil {
		_ = json.Unmarshal(m.Facts, &r.Facts)
	}
}

//
// Model builds a model.
func (r *AnalysisIncident) Model() (m *model.AnalysisIncident) {
	m = &model.AnalysisIncident{}
	m.URI = r.URI
	m.Message = r.Message
	m.Facts, _ = json.Marshal(r.Facts)
	return
}

//
// AnalysisLink analysis report link.
type AnalysisLink struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty" yaml:",omitempty"`
}

//
// IssueComposite composite REST resource.
type IssueComposite struct {
	RuleID      string   `json:"ruleID"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Effort      int      `json:"effort"`
	Labels      []string `json:"labels"`
	Affected    int      `json:"affected"`
}

//
// DepComposite composite REST resource.
type DepComposite struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	SHA      string `json:"sha"`
	Affected int    `json:"affected"`
}

//
// FactMap map.
type FactMap map[string]interface{}
