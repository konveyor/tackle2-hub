package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strings"
	"time"
)

//
// Routes
const (
	AnalysesRoot               = "/analyses"
	AnalysisRoot               = AnalysesRoot + "/:" + ID
	AnalysesDepsRoot           = AnalysesRoot + "/dependencies"
	AnalysesIssuesRoot         = AnalysesRoot + "/issues"
	AnalysesCompositeRoot      = AnalysesRoot + "/composite"
	AnalysisCompoSiteDepRoot   = AnalysesCompositeRoot + "/dependencies"
	AnalysisCompoSiteIssueRoot = AnalysesCompositeRoot + "/issues"

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
	routeGroup.GET(AnalysisCompoSiteIssueRoot, h.IssueComposites)
	routeGroup.GET(AnalysisCompoSiteDepRoot, h.DepComposites)
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
		"RuleSets.Issues",
		"RuleSets.Issues.Incidents",
		"RuleSets.Technologies")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Analysis{}
	r.With(m)

	h.Render(ctx, http.StatusOK, r)
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
		"RuleSets.Issues",
		"RuleSets.Issues.Incidents",
		"RuleSets.Technologies")
	db = db.Where("ApplicationID = ?", id)
	result := db.Last(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Analysis{}
	r.With(m)

	h.Render(ctx, http.StatusOK, r)
}

// AppList godoc
// @summary List analyses.
// @description List analyses for an application.
// @description Resources do not include relations.
// @tags analyses
// @produce json
// @success 200 {object} []api.Analyses
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
		h.Render(ctx, http.StatusOK, resources)
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

	h.Render(ctx, http.StatusOK, resources)
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
	mark := time.Now()
	m := r.Model()
	m.ApplicationID = id
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	rtx := WithContext(ctx)
	result := rtx.DB.Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

	Log.Info(ctx.Request.URL.String(), "duration", time.Since(mark))

	h.Render(ctx, http.StatusCreated, r)
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

	ctx.Status(http.StatusNoContent)
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
// @success 200 {object} []api.AnalysesDependency
// @router /application/{id}/analysis/dependencies [get]
// @param id path string true "Application ID"
func (h AnalysisHandler) AppDeps(ctx *gin.Context) {
	resources := []AnalysisDependency{}
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
	db = h.Paginated(ctx)
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
		h.Render(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
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

	h.Render(ctx, http.StatusOK, resources)
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
// @success 200 {object} []api.AnalysesIssue
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
	db = h.Paginated(ctx)
	ruleSet := h.DB(ctx)
	ruleSet = ruleSet.Model(&model.AnalysisRuleSet{})
	ruleSet = ruleSet.Select("ID")
	ruleSet = ruleSet.Where("AnalysisID = ?", analysis.ID)
	db = db.Where("RuleSetID IN (?)", ruleSet)
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result = db.Model(&model.AnalysisIssue{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Render(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	list := []model.AnalysisIssue{}
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

	h.Render(ctx, http.StatusOK, resources)
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
// @description - application.(id|name)
// @description - tech.(source|target)
// @description - tag.id
// @tags issues
// @produce json
// @success 200 {object} []api.AnalysesIssue
// @router /analyses/issues [get]
func (h AnalysisHandler) Issues(ctx *gin.Context) {
	resources := []AnalysisIssue{}
	mark := time.Now()
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "ruleid", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "affected", Kind: qf.LITERAL},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tech.source", Kind: qf.STRING},
			{Field: "tech.target", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, Relation: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.Paginated(ctx)
	db = db.Table("AnalysisIssue i,")
	db = db.Joins("AnalysisRuleSet r,")
	db = db.Joins("Analysis a")
	db = db.Where("a.ID = r.AnalysisID")
	db = db.Where("r.ID = i.RuleSetID")
	db = db.Where("r.ID IN (?)", h.rulesetIDs(ctx, &filter))
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result := db.Model(&model.AnalysisIssue{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Render(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	type M struct {
		model.AnalysisIssue
		ApplicationID uint
	}
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

	Log.Info(ctx.Request.URL.String(), "duration", time.Since(mark))

	h.Render(ctx, http.StatusOK, resources)
}

// IssueComposites godoc
// @summary List issue composites.
// @description List issue composites.
// @description filters:
// @description - category
// @description - effort
// @description - application.(id|name)
// @description - tech.(source|target)
// @description - tag.id
// @tags issuecomposites
// @produce json
// @success 200 {object} []api.IssueComposite
// @router /analyses/issues [get]
func (h AnalysisHandler) IssueComposites(ctx *gin.Context) {
	resources := []*IssueComposite{}
	mark := time.Now()
	// Build query.
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "affected", Kind: qf.LITERAL},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tech.source", Kind: qf.STRING},
			{Field: "tech.target", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, Relation: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	p := Page{}
	p.With(ctx)
	sort := Sort{}
	sort.With(ctx)
	// Build query.
	ruleSets := h.rulesetIDs(ctx, &filter)
	q := h.DB(ctx)
	q = q.Select(
		"i.RuleID",
		"i.Category", // needed by filter
		"i.Effort",   // needed by filter
		"COUNT(a.ID) Affected")
	q = q.Table("AnalysisIssue i,")
	q = q.Joins("AnalysisRuleSet r,")
	q = q.Joins("Analysis a")
	q = q.Where("a.ID = r.AnalysisID")
	q = q.Where("r.ID = i.RulesetID")
	q = q.Where("r.ID IN (?)", ruleSets)
	q = q.Group("i.RuleID")
	q = q.Order("i.RuleID")
	db := h.DB(ctx)
	db = db.Table("(?)", q)
	db = filter.Where(db)
	// Count.
	count := int64(0)
	result := db.Model(&model.AnalysisIssue{}).Count(&count)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if count == 0 {
		h.Render(ctx, http.StatusOK, resources)
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
		RuleID      string
		RuleName    string
		Description string
		Category    string
		Effort      int
		Labels      model.JSON
		Name        string
		Version     string
		Source      bool
		Affected    int
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
	q = p.Paginated(q)
	db = h.DB(ctx)
	db = db.Select(
		"i.RuleID",
		"i.Name RuleName",
		"i.Description",
		"i.Category",
		"i.Effort",
		"i.Labels",
		"t.Name",
		"t.Version",
		"t.Source")
	db = sort.Sorted(db)
	db = db.Table("AnalysisIssue i,")
	db = db.Joins("AnalysisRuleSet r,")
	db = db.Joins("AnalysisTechnology t,")
	db = db.Joins("Analysis a")
	db = db.Where("a.ID = r.AnalysisID")
	db = db.Where("r.ID = i.RulesetID")
	db = db.Where("r.ID IN (?)", ruleSets)
	db = db.Where("t.RulesetID = r.ID")
	db = db.Where("i.RuleID IN (SELECT RuleID from (?))", q)
	db = filter.Where(db)
	result = db.Scan(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	collated := make(map[string]*IssueComposite)
	for i := range list {
		m := list[i]
		r, found := collated[m.RuleID]
		if !found {
			r = &IssueComposite{
				tech:        make(map[string]AnalysisTechnology),
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
		tech := AnalysisTechnology{
			Name:    m.Name,
			Version: m.Version,
			Source:  m.Source,
		}
		r.tech[tech.key()] = tech
	}
	for _, r := range resources {
		for _, tech := range r.tech {
			r.Technologies = append(
				r.Technologies,
				tech)
		}
	}

	Log.Info(ctx.Request.URL.String(), "duration", time.Since(mark))

	h.Render(ctx, http.StatusOK, resources)
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
// @success 200 {object} []api.AnalysesDependency
// @router /analyses/dependencies [get]
func (h AnalysisHandler) Deps(ctx *gin.Context) {
	resources := []AnalysisDependency{}
	mark := time.Now()
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
	db := h.Paginated(ctx)
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
		h.Render(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
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

	Log.Info(ctx.Request.URL.String(), "duration", time.Since(mark))

	h.Render(ctx, http.StatusOK, resources)
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
// @success 200 {object} []api.AnalysesDependency
// @router /analyses/dependencies [get]
func (h AnalysisHandler) DepComposites(ctx *gin.Context) {
	resources := []DepComposite{}
	mark := time.Now()
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
	db := h.Paginated(ctx)
	db = db.Where("AnalysisID IN (?)", h.analysisIDs(ctx, &filter))
	db = filter.Where(db)
	db = db.Select(
		"Name",
		"Version",
		"Type",
		"SHA",
		"COUNT(AnalysisID) Affected")
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
		h.Render(ctx, http.StatusOK, resources)
		return
	}
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	result = db.Find(&resources)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	Log.Info(ctx.Request.URL.String(), "duration", time.Since(mark))

	h.Render(ctx, http.StatusOK, resources)
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
			part := []string{}
			values := []interface{}{}
			for i, v := range field.Value.ByKind(qf.LITERAL, qf.STRING) {
				values = append(values, qf.AsValue(v))
				if i > 0 {
					part = append(part, "INTERSECT")
				}
				part = append(
					part,
					"SELECT",
					"  applicationID",
					"FROM",
					"  ApplicationTags",
					"WHERE",
					"  TagID = ?")
			}
			tags := h.DB(ctx).Raw(
				strings.Join(part, " "),
				values...)
			q = q.Where("ID IN (?)", tags)
		} else {
			field = field.As("TagID")
			tags := h.DB(ctx)
			tags = tags.Model(&model.ApplicationTag{})
			tags = tags.Select("ApplicationID")
			tags = tags.Where(field.SQL())
			q = q.Where("ID IN (?)", tags)
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
// rulesetIDs provides ruleSet IDs.
// filter:
//   - tech.source
//   - tech.target
func (h *AnalysisHandler) rulesetIDs(ctx *gin.Context, f *qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.AnalysisRuleSet{})
	q = q.Select("ID")
	q = q.Where("AnalysisID IN (?)", h.analysisIDs(ctx, f))
	techFilter := f.Resource("tech")
	if field, found := techFilter.Field("source"); found {
		field = field.As("Name")
		tech := h.DB(ctx)
		tech = tech.Model(&model.AnalysisTechnology{})
		tech = tech.Select("RuleSetID")
		tech = tech.Where("Source", true)
		tech = tech.Where(field.SQL())
		q = q.Where("ID IN (?)", tech)
	}
	if field, found := techFilter.Field("target"); found {
		field = field.As("Name")
		tech := h.DB(ctx)
		tech = tech.Model(&model.AnalysisTechnology{})
		tech = tech.Select("RuleSetID")
		tech = tech.Where("Source", false)
		tech = tech.Where(field.SQL())
		q = q.Where("ID IN (?)", tech)
	}
	return
}

//
// Analysis (Analysis) REST resource.
type Analysis struct {
	Resource     `yaml:",inline"`
	RuleSets     []AnalysisRuleSet    `json:"ruleSets"`
	Dependencies []AnalysisDependency `json:"dependencies"`
}

//
// With updates the resource with the model.
func (r *Analysis) With(m *model.Analysis) {
	r.Resource.With(&m.Model)
	r.RuleSets = []AnalysisRuleSet{}
	for i := range m.RuleSets {
		n := AnalysisRuleSet{}
		n.With(&m.RuleSets[i])
		r.RuleSets = append(
			r.RuleSets,
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
	m.RuleSets = []model.AnalysisRuleSet{}
	for i := range r.RuleSets {
		n := r.RuleSets[i].Model()
		m.RuleSets = append(
			m.RuleSets,
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
// AnalysisRuleSet REST resource.
type AnalysisRuleSet struct {
	Resource     `yaml:",inline"`
	Name         string               `json:"name" binding:"required"`
	Description  string               `json:"description"`
	Technologies []AnalysisTechnology `json:"technologies"`
	Issues       []AnalysisIssue      `json:"issues"`
}

//
// With updates the resource with the model.
func (r *AnalysisRuleSet) With(m *model.AnalysisRuleSet) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Technologies = []AnalysisTechnology{}
	for i := range m.Technologies {
		n := AnalysisTechnology{}
		n.With(&m.Technologies[i])
		r.Technologies = append(
			r.Technologies,
			n)
	}
	r.Issues = []AnalysisIssue{}
	for i := range m.Issues {
		n := AnalysisIssue{}
		n.With(&m.Issues[i])
		r.Issues = append(
			r.Issues,
			n)
	}
}

//
// Model builds a model.
func (r *AnalysisRuleSet) Model() (m *model.AnalysisRuleSet) {
	m = &model.AnalysisRuleSet{}
	m.Name = r.Name
	m.Description = r.Description
	m.Technologies = []model.AnalysisTechnology{}
	for i := range r.Technologies {
		n := r.Technologies[i].Model()
		m.Technologies = append(
			m.Technologies,
			*n)
	}
	m.Issues = []model.AnalysisIssue{}
	for i := range r.Issues {
		n := r.Issues[i].Model()
		m.Issues = append(
			m.Issues,
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
// AnalysisTechnology REST resource.
type AnalysisTechnology struct {
	Resource `yaml:",inline"`
	Name     string `json:"name" binding:"required"`
	Version  string `json:"version,omitempty" yaml:",omitempty"`
	Source   bool   `json:"source,omitempty" yaml:",omitempty"`
}

//
// With updates the resource with the model.
func (r *AnalysisTechnology) With(m *model.AnalysisTechnology) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Version = m.Version
	r.Source = m.Source
}

//
// key returns a unique key.
func (r *AnalysisTechnology) key() string {
	return fmt.Sprintf(
		"%s:%s:%v",
		r.Name,
		r.Version,
		r.Source)
}

//
// Model builds a model.
func (r *AnalysisTechnology) Model() (m *model.AnalysisTechnology) {
	m = &model.AnalysisTechnology{}
	m.Name = r.Name
	m.Version = r.Version
	m.Source = r.Source
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
	tech         map[string]AnalysisTechnology
	RuleID       string               `json:"ruleID"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Category     string               `json:"category"`
	Effort       int                  `json:"effort"`
	Labels       []string             `json:"labels"`
	Technologies []AnalysisTechnology `json:"technologies"`
	Affected     int                  `json:"affected"`
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
