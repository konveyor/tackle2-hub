package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/tar"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Routes
const (
	AnalysesRoot          = "/analyses"
	AnalysisRoot          = AnalysesRoot + "/:" + ID
	AnalysesDepsRoot      = AnalysesRoot + "/dependencies"
	AnalysesIssuesRoot    = AnalysesRoot + "/issues"
	AnalysesIssueRoot     = AnalysesIssuesRoot + "/:" + ID
	AnalysisIncidentsRoot = AnalysesIssueRoot + "/incidents"
	//
	AnalysesReportRoot           = AnalysesRoot + "/report"
	AnalysisReportDepsRoot       = AnalysesReportRoot + "/dependencies"
	AnalysisReportRuleRoot       = AnalysesReportRoot + "/rules"
	AnalysisReportIssuesRoot     = AnalysesReportRoot + "/issues"
	AnalysisReportAppsRoot       = AnalysesReportRoot + "/applications"
	AnalysisReportIssueRoot      = AnalysisReportIssuesRoot + "/:" + ID
	AnalysisReportIssuesAppsRoot = AnalysisReportIssuesRoot + "/applications"
	AnalysisReportDepsAppsRoot   = AnalysisReportDepsRoot + "/applications"
	AnalysisReportAppsIssuesRoot = AnalysisReportAppsRoot + "/:" + ID + "/issues"
	AnalysisReportFileRoot       = AnalysisReportIssueRoot + "/files"
	//
	AppAnalysesRoot       = ApplicationRoot + "/analyses"
	AppAnalysisRoot       = ApplicationRoot + "/analysis"
	AppAnalysisReportRoot = AppAnalysisRoot + "/report"
	AppAnalysisDepsRoot   = AppAnalysisRoot + "/dependencies"
	AppAnalysisIssuesRoot = AppAnalysisRoot + "/issues"
)

const (
	IssueField = "issues"
	DepField   = "dependencies"
)

// AnalysisHandler handles analysis resource routes.
type AnalysisHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h AnalysisHandler) AddRoutes(e *gin.Engine) {
	// Primary
	routeGroup := e.Group("/")
	routeGroup.Use(Required("analyses"))
	routeGroup.GET(AnalysisRoot, h.Get)
	routeGroup.DELETE(AnalysisRoot, h.Delete)
	routeGroup.GET(AnalysesDepsRoot, h.Deps)
	routeGroup.GET(AnalysesIssuesRoot, h.Issues)
	routeGroup.GET(AnalysesIssueRoot, h.Issue)
	routeGroup.GET(AnalysisIncidentsRoot, h.Incidents)
	routeGroup.GET(AnalysisReportRuleRoot, h.RuleReports)
	routeGroup.GET(AnalysisReportAppsIssuesRoot, h.AppIssueReports)
	routeGroup.GET(AnalysisReportIssuesAppsRoot, h.IssueAppReports)
	routeGroup.GET(AnalysisReportFileRoot, h.FileReports)
	routeGroup.GET(AnalysisReportDepsRoot, h.DepReports)
	routeGroup.GET(AnalysisReportDepsAppsRoot, h.DepAppReports)
	// Application
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.analyses"))
	routeGroup.POST(AppAnalysesRoot, h.AppCreate)
	routeGroup.GET(AppAnalysesRoot, h.AppList)
	routeGroup.GET(AppAnalysisRoot, h.AppLatest)
	routeGroup.GET(AppAnalysisReportRoot, h.AppLatestReport)
	routeGroup.GET(AppAnalysisDepsRoot, h.AppDeps)
	routeGroup.GET(AppAnalysisIssuesRoot, h.AppIssues)
}

// Get godoc
// @summary Get an analysis (report) by ID.
// @description Get an analysis (report) by ID.
// @tags analyses
// @produce octet-stream
// @success 200 {object} api.Analysis
// @router /analyses/{id} [get]
// @param id path int true "Analysis ID"
func (h AnalysisHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	writer := AnalysisWriter{ctx: ctx}
	path, err := writer.Create(id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = os.Remove(path)
	}()
	h.Status(ctx, http.StatusOK)
	ctx.File(path)
}

// AppLatest godoc
// @summary Get the latest analysis.
// @description Get the latest analysis for an application.
// @tags analyses
// @produce octet-stream
// @success 200 {object} api.Analysis
// @router /applications/{id}/analysis [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppLatest(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Analysis{}
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	err := db.Last(&m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	writer := AnalysisWriter{ctx: ctx}
	path, err := writer.Create(m.ID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = os.Remove(path)
	}()
	h.Status(ctx, http.StatusOK)
	ctx.File(path)
}

// AppLatestReport godoc
// @summary Get the latest analysis (static) report.
// @description Get the latest analysis (static) report.
// @tags analyses
// @produce octet-stream
// @success 200
// @router /applications/{id}/analysis/report [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppLatestReport(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Analysis{}
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	err := db.Last(&m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	reportWriter := ReportWriter{ctx: ctx}
	reportWriter.Write(m.ID)
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
	// Sort
	sort := Sort{}
	err := sort.With(ctx, &model.Issue{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find.
	id := h.pk(ctx)
	db := h.DB(ctx)
	db = db.Model(&model.Analysis{})
	db = db.Where("ApplicationID = ?", id)
	db = sort.Sorted(db)
	var list []model.Analysis
	var m model.Analysis
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
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
// @description Form fields:
// @description   - file: file that contains the api.Analysis resource.
// @description   - issues: file that multiple api.Issue resources.
// @description   - dependencies: file that multiple api.TechDependency resources.
// @tags analyses
// @produce json
// @success 201 {object} api.Analysis
// @router /application/{id}/analyses [post]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppCreate(ctx *gin.Context) {
	id := h.pk(ctx)
	result := h.DB(ctx).First(&model.Application{}, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err := h.archive(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	analysis := &model.Analysis{}
	analysis.ApplicationID = id
	analysis.CreateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx)
	db.Logger = db.Logger.LogMode(logger.Error)
	err = db.Create(analysis).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Analysis
	input, err := ctx.FormFile(FileField)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	reader, err := input.Open()
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	encoding := input.Header.Get(ContentType)
	d, err := h.Decoder(ctx, encoding, reader)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	r := Analysis{}
	err = d.Decode(&r)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	//
	// Issues
	input, err = ctx.FormFile(IssueField)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	reader, err = input.Open()
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	encoding = input.Header.Get(ContentType)
	d, err = h.Decoder(ctx, encoding, reader)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	for {
		r := &Issue{}
		err = d.Decode(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				err = &BadRequestError{err.Error()}
				_ = ctx.Error(err)
				return
			}
		}
		m := r.Model()
		m.AnalysisID = analysis.ID
		err = db.Create(m).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		analysis.Effort += r.Effort * len(r.Incidents)
	}
	//
	// Dependencies
	input, err = ctx.FormFile(DepField)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	reader, err = input.Open()
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	encoding = input.Header.Get(ContentType)
	d, err = h.Decoder(ctx, encoding, reader)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	var deps []*TechDependency
	for {
		r := &TechDependency{}
		err = d.Decode(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				err = &BadRequestError{err.Error()}
				_ = ctx.Error(err)
				return
			}
		}
		deps = append(deps, r)
	}
	sort.Slice(deps, func(i, _ int) bool {
		return !deps[i].Indirect
	})
	for _, r := range deps {
		m := r.Model()
		m.AnalysisID = analysis.ID
		db := db.Clauses(clause.OnConflict{DoNothing: true})
		err = db.Create(m).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}
	//
	// Update effort.
	err = db.Save(analysis).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	db = h.DB(ctx)
	db = db.Preload(clause.Associations)
	err = db.First(analysis).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r.With(analysis)

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an analysis by ID.
// @description Delete an analysis by ID.
// @tags analyses
// @success 204
// @router /analyses/{id} [delete]
// @param id path int true "Analysis ID"
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
// @description - name
// @description - version
// @description - sha
// @description - indirect
// @description - labels
// @tags dependencies
// @produce json
// @success 200 {object} []api.TechDependency
// @router /application/{id}/analysis/dependencies [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppDeps(ctx *gin.Context) {
	resources := []TechDependency{}
	// Latest
	id := h.pk(ctx)
	analysis := &model.Analysis{}
	db := h.DB(ctx)
	db = db.Where("ApplicationID = ?", id)
	result := db.Last(analysis)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "name", Kind: qf.STRING},
			{Field: "version", Kind: qf.STRING},
			{Field: "sha", Kind: qf.STRING},
			{Field: "indirect", Kind: qf.STRING},
			{Field: "labels", Kind: qf.STRING, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	sort := Sort{}
	err = sort.With(ctx, &model.TechDependency{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	db = h.DB(ctx)
	db = db.Model(&model.TechDependency{})
	db = db.Where("AnalysisID = ?", analysis.ID)
	db = db.Where("ID IN (?)", h.depIDs(ctx, filter))
	db = sort.Sorted(db)
	var list []model.TechDependency
	var m model.TechDependency
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		r := TechDependency{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AppIssues godoc
// @summary List application issues.
// @description List application issues.
// @description filters:
// @description - ruleset
// @description - rule
// @description - name
// @description - category
// @description - effort
// @description - labels
// @tags issues
// @produce json
// @success 200 {object} []api.Issue
// @router /application/{id}/analysis/issues [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppIssues(ctx *gin.Context) {
	resources := []Issue{}
	// Latest
	id := h.pk(ctx)
	analysis := &model.Analysis{}
	db := h.DB(ctx).Where("ApplicationID = ?", id)
	result := db.Last(analysis)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "ruleset", Kind: qf.STRING},
			{Field: "rule", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "labels", Kind: qf.STRING, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &model.Issue{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	db = h.DB(ctx)
	db = db.Model(&model.Issue{})
	db = db.Where("AnalysisID = ?", analysis.ID)
	db = db.Where("ID IN (?)", h.issueIDs(ctx, filter))
	db = sort.Sorted(db)
	var list []model.Issue
	var m model.Issue
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := &list[i]
		r := Issue{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Issues godoc
// @summary List all issues.
// @description List all issues.
// @description filters:
// @description - ruleset
// @description - rule
// @description - name
// @description - category
// @description - effort
// @description - labels
// @description - application.id
// @description - application.name
// @description - tag.id
// @tags issues
// @produce json
// @success 200 {object} []api.Issue
// @router /analyses/issues [get]
func (h AnalysisHandler) Issues(ctx *gin.Context) {
	resources := []Issue{}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "ruleset", Kind: qf.STRING},
			{Field: "rule", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "labels", Kind: qf.STRING, And: true},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &model.Issue{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	db := h.DB(ctx)
	db = db.Table("Issue i")
	db = db.Joins(",Analysis a")
	db = db.Where("a.ID = i.AnalysisID")
	db = db.Where("a.ID IN (?)", h.analysisIDs(ctx, filter))
	db = db.Where("i.ID IN (?)", h.issueIDs(ctx, filter))
	db = db.Group("i.ID")
	db = sort.Sorted(db)
	var list []model.Issue
	var m model.Issue
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := &list[i]
		r := Issue{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Issue godoc
// @summary Get an issue.
// @description Get an issue.
// @tags issue
// @produce json
// @success 200 {object} api.Issue
// @router /analyses/issues/{id} [get]
// @param id path int true "Issue ID"
func (h AnalysisHandler) Issue(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Issue{}
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Issue{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// Incidents godoc
// @summary List incidents for an issue.
// @description List incidents for an issue.
// @description filters:
// @description - file
// @tags incidents
// @produce json
// @success 200 {object} []api.Incident
// @router /analyses/issues/{id}/incidents [get]
// @param id path int true "Issue ID"
func (h AnalysisHandler) Incidents(ctx *gin.Context) {
	issueId := ctx.Param(ID)
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "file", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &model.Incident{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	db := h.DB(ctx)
	db = db.Model(&model.Incident{})
	db = db.Where("IssueID", issueId)
	db = filter.Where(db)
	db = sort.Sorted(db)
	var list []model.Incident
	var m model.Incident
	cursor := Cursor{}
	defer func() {
		cursor.Close()
	}()
	page := Page{}
	page.With(ctx)
	cursor.With(db, page)
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	resources := []Incident{}
	for _, m := range list {
		r := Incident{}
		r.With(&m)
		resources = append(
			resources,
			r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// RuleReports godoc
// @summary List rule reports.
// @description Each report collates issues by ruleset/rule.
// @description filters:
// @description - ruleset
// @description - rule
// @description - category
// @description - effort
// @description - labels
// @description - applications
// @description - application.id
// @description - application.name
// @description - businessService.id
// @description - businessService.name
// @description - tag.id
// @description sort:
// @description - ruleset
// @description - rule
// @description - category
// @description - effort
// @description - applications
// @tags rulereports
// @produce json
// @success 200 {object} []api.RuleReport
// @router /analyses/report/rules [get]
func (h AnalysisHandler) RuleReports(ctx *gin.Context) {
	resources := []*RuleReport{}
	type M struct {
		model.Issue
		Applications int
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "ruleset", Kind: qf.STRING},
			{Field: "rule", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "labels", Kind: qf.STRING, And: true},
			{Field: "applications", Kind: qf.LITERAL},
			{Field: "application.id", Kind: qf.STRING},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "businessService.id", Kind: qf.LITERAL},
			{Field: "businessService.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &M{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Inner Query
	q := h.DB(ctx)
	q = q.Select(
		"i.RuleSet",
		"i.Rule",
		"i.Name",
		"i.Description",
		"i.Category",
		"i.Effort",
		"i.Labels",
		"i.Links",
		"COUNT(distinct a.ID) Applications")
	q = q.Table("Issue i,")
	q = q.Joins("Analysis a")
	q = q.Where("a.ID = i.AnalysisID")
	q = q.Where("a.ID in (?)", h.analysisIDs(ctx, filter))
	q = q.Where("i.ID IN (?)", h.issueIDs(ctx, filter))
	q = q.Group("i.RuleSet,i.Rule")
	// Find
	db := h.DB(ctx)
	db = db.Select("*")
	db = db.Table("(?)", q)
	db = sort.Sorted(db)
	var list []M
	var m M
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := list[i]
		r := &RuleReport{
			Applications: m.Applications,
			Description:  m.Description,
			Category:     m.Category,
			RuleSet:      m.RuleSet,
			Rule:         m.Rule,
			Name:         m.Name,
		}
		resources = append(resources, r)
		if m.Labels != nil {
			_ = json.Unmarshal(m.Labels, &r.Labels)
		}
		if m.Links != nil {
			_ = json.Unmarshal(m.Links, &r.Links)
		}
		r.Effort += m.Effort
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AppIssueReports godoc
// @summary List application issue reports.
// @description Each report collates issues by ruleset/rule.
// @description filters:
// @description - ruleset
// @description - rule
// @description - category
// @description - effort
// @description - labels
// @description sort:
// @description - ruleset
// @description - rule
// @description - category
// @description - effort
// @description - files
// @tags issuereport
// @produce json
// @success 200 {object} []api.IssueReport
// @router /analyses/report/applications/{id}/issues [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppIssueReports(ctx *gin.Context) {
	resources := []*IssueReport{}
	type M struct {
		model.Issue
		Files int
	}
	id := h.pk(ctx)
	err := h.DB(ctx).First(&model.Application{}, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Latest
	analysis := &model.Analysis{}
	db := h.DB(ctx).Where("ApplicationID", id)
	err = db.Last(analysis).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.Respond(
				ctx,
				http.StatusOK,
				resources)
		} else {
			_ = ctx.Error(err)
		}
		return
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "ruleset", Kind: qf.STRING},
			{Field: "rule", Kind: qf.STRING},
			{Field: "category", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "labels", Kind: qf.STRING, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &M{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Inner Query
	q := h.DB(ctx)
	q = q.Select(
		"i.ID",
		"i.RuleSet",
		"i.Rule",
		"i.Name",
		"i.Description",
		"i.Category",
		"i.Effort",
		"i.Labels",
		"i.Links",
		"COUNT(distinct n.File) Files")
	q = q.Table("Issue i,")
	q = q.Joins("Incident n")
	q = q.Where("i.ID = n.IssueID")
	q = q.Where("i.ID IN (?)", h.issueIDs(ctx, filter))
	q = q.Where("i.AnalysisID", analysis.ID)
	q = q.Group("i.RuleSet,i.Rule")
	// Find
	db = h.DB(ctx)
	db = db.Select("*")
	db = db.Table("(?)", q)
	db = sort.Sorted(db)
	var list []M
	var m M
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := list[i]
		r := &IssueReport{
			Files:       m.Files,
			Description: m.Description,
			Category:    m.Category,
			RuleSet:     m.RuleSet,
			Rule:        m.Rule,
			Name:        m.Name,
			ID:          m.ID,
		}
		resources = append(resources, r)
		if m.Labels != nil {
			_ = json.Unmarshal(m.Labels, &r.Labels)
		}
		if m.Links != nil {
			_ = json.Unmarshal(m.Links, &r.Links)
		}
		r.Effort += m.Effort
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// IssueAppReports godoc
// @summary List application reports.
// @description List application reports.
// @description filters:
// @description - id
// @description - name
// @description - description
// @description - businessService
// @description - effort
// @description - incidents
// @description - files
// @description - issue.id
// @description - issue.name
// @description - issue.ruleset
// @description - issue.rule
// @description - issue.category
// @description - issue.effort
// @description - issue.labels
// @description - application.id
// @description - application.name
// @description - businessService.id
// @description - businessService.name
// @description sort:
// @description - id
// @description - name
// @description - description
// @description - businessService
// @description - effort
// @description - incidents
// @description - files
// @tags issueappreports
// @produce json
// @success 200 {object} []api.IssueAppReport
// @router /analyses/report/applications [get]
func (h AnalysisHandler) IssueAppReports(ctx *gin.Context) {
	resources := []IssueAppReport{}
	type M struct {
		ID               uint
		Name             string
		Description      string
		BusinessService  string
		Effort           int
		Incidents        int
		Files            int
		IssueID          uint
		IssueName        string
		IssueDescription string
		RuleSet          string
		Rule             string
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "description", Kind: qf.STRING},
			{Field: "businessService", Kind: qf.STRING},
			{Field: "effort", Kind: qf.LITERAL},
			{Field: "incidents", Kind: qf.LITERAL},
			{Field: "files", Kind: qf.LITERAL},
			{Field: "issue.id", Kind: qf.LITERAL},
			{Field: "issue.name", Kind: qf.LITERAL},
			{Field: "issue.ruleset", Kind: qf.STRING},
			{Field: "issue.rule", Kind: qf.STRING},
			{Field: "issue.category", Kind: qf.STRING},
			{Field: "issue.effort", Kind: qf.LITERAL},
			{Field: "issue.labels", Kind: qf.STRING, And: true},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "businessService.id", Kind: qf.LITERAL},
			{Field: "businessService.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &M{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Inner Query
	q := h.DB(ctx)
	q = q.Select(
		"app.ID",
		"app.Name",
		"app.Description",
		"b.Name BusinessService",
		"a.Effort",
		"COUNT(n.ID) Incidents",
		"COUNT(distinct n.File) Files",
		"i.ID IssueID",
		"i.Name IssueName",
		"i.RuleSet",
		"i.Rule")
	q = q.Table("Issue i")
	q = q.Joins("LEFT JOIN Incident n ON n.IssueID = i.ID")
	q = q.Joins("LEFT JOIN Analysis a ON a.ID = i.AnalysisID")
	q = q.Joins("LEFT JOIN Application app ON app.ID = a.ApplicationID")
	q = q.Joins("LEFT OUTER JOIN BusinessService b ON b.ID = app.BusinessServiceID")
	q = q.Where("a.ID IN (?)", h.analysisIDs(ctx, filter))
	q = q.Where("i.ID IN (?)", h.issueIDs(ctx, filter.Resource("issue")))
	q = q.Group("i.ID")
	// Find
	db := h.DB(ctx)
	db = db.Select("*")
	db = db.Table("(?)", q)
	db = filter.Where(db)
	db = sort.Sorted(db)
	var list []M
	var m M
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := &list[i]
		r := IssueAppReport{}
		r.ID = m.ID
		r.Name = m.Name
		r.Description = m.Description
		r.BusinessService = m.BusinessService
		r.Effort = m.Effort
		r.Incidents = m.Incidents
		r.Files = m.Files
		r.Issue.ID = m.IssueID
		r.Issue.Name = m.IssueName
		r.Issue.Description = m.IssueName
		r.Issue.RuleSet = m.RuleSet
		r.Issue.Rule = m.Rule
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// FileReports godoc
// @summary List incident file reports.
// @description Each report collates incidents by file.
// @description filters:
// @description - file
// @description - effort
// @description - incidents
// @description sort:
// @description - file
// @description - effort
// @description - incidents
// @tags filereports
// @produce json
// @success 200 {object} []api.FileReport
// @router /analyses/report/issues/{id}/files [get]
// @param id path int true "Issue ID"
func (h AnalysisHandler) FileReports(ctx *gin.Context) {
	resources := []FileReport{}
	type M struct {
		IssueId   uint
		File      string
		Effort    int
		Incidents int
	}
	// Issue
	issueId := h.pk(ctx)
	issue := &model.Issue{}
	result := h.DB(ctx).First(issue, issueId)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "file", Kind: qf.STRING},
			{Field: "incidents", Kind: qf.LITERAL},
			{Field: "effort", Kind: qf.LITERAL},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &M{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Inner Query
	q := h.DB(ctx)
	q = q.Model(&model.Incident{})
	q = q.Select(
		"IssueId",
		"File",
		"Effort*COUNT(Incident.id) Effort",
		"COUNT(Incident.id) Incidents")
	q = q.Joins(",Issue")
	q = q.Where("Issue.ID = IssueID")
	q = q.Where("Issue.ID", issueId)
	q = q.Group("File")
	// Find
	db := h.DB(ctx)
	db = db.Select("*")
	db = db.Table("(?)", q)
	db = filter.Where(db)
	db = sort.Sorted(db)
	var list []M
	var m M
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for _, m := range list {
		r := FileReport{}
		r.IssueID = m.IssueId
		r.File = m.File
		r.Effort = m.Effort
		r.Incidents = m.Incidents
		resources = append(
			resources,
			r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Deps godoc
// @summary List dependencies.
// @description List dependencies.
// @description filters:
// @description - name
// @description - version
// @description - sha
// @description - indirect
// @description - labels
// @description - application.id
// @description - application.name
// @description - tag.id
// @tags dependencies
// @produce json
// @success 200 {object} []api.TechDependency
// @router /analyses/dependencies [get]
func (h AnalysisHandler) Deps(ctx *gin.Context) {
	resources := []TechDependency{}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "name", Kind: qf.STRING},
			{Field: "version", Kind: qf.STRING},
			{Field: "sha", Kind: qf.STRING},
			{Field: "indirect", Kind: qf.STRING},
			{Field: "labels", Kind: qf.STRING, And: true},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &model.TechDependency{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	db := h.DB(ctx)
	db = db.Model(&model.TechDependency{})
	db = db.Where("AnalysisID IN (?)", h.analysisIDs(ctx, filter))
	db = db.Where("ID IN (?)", h.depIDs(ctx, filter))
	db = sort.Sorted(db)
	var list []model.TechDependency
	var m model.TechDependency
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		r := TechDependency{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// DepReports godoc
// @summary List dependency reports.
// @description Each report collates dependencies by name and SHA.
// @description filters:
// @description - provider
// @description - name
// @description - version
// @description - sha
// @description - indirect
// @description - labels
// @description - application.id
// @description - application.name
// @description - businessService.id
// @description - businessService.name
// @description - tag.id
// @description sort:
// @description - provider
// @description - name
// @description - labels
// @tags dependencies
// @produce json
// @success 200 {object} []api.TechDependency
// @router /analyses/dependencies [get]
func (h AnalysisHandler) DepReports(ctx *gin.Context) {
	resources := []DepReport{}
	type M struct {
		model.TechDependency
		Applications int
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "provider", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "version", Kind: qf.STRING},
			{Field: "sha", Kind: qf.STRING},
			{Field: "indirect", Kind: qf.STRING},
			{Field: "labels", Kind: qf.STRING, And: true},
			{Field: "applications", Kind: qf.LITERAL},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "businessService.id", Kind: qf.LITERAL},
			{Field: "businessService.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &M{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Inner Query
	q := h.DB(ctx)
	q = q.Select(
		"d.Provider",
		"d.Name",
		"json_group_array(distinct j.value) Labels",
		"COUNT(distinct d.AnalysisID) Applications")
	q = q.Table("TechDependency d")
	q = q.Joins(",json_each(Labels) j")
	q = q.Where("d.AnalysisID IN (?)", h.analysisIDs(ctx, filter))
	q = q.Where("d.ID IN (?)", h.depIDs(ctx, filter))
	q = q.Group("d.Provider, d.Name")
	// Find
	db := h.DB(ctx)
	db = db.Select("*")
	db = db.Table("(?)", q)
	db = sort.Sorted(db)
	var list []M
	var m M
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := &list[i]
		r := DepReport{
			Provider:     m.Provider,
			Name:         m.Name,
			Applications: m.Applications,
		}
		if m.Labels != nil {
			var aggregated []string
			_ = json.Unmarshal(m.Labels, &aggregated)
			for _, s := range aggregated {
				if s != "" {
					r.Labels = append(r.Labels, s)
				}
			}
		}
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// DepAppReports godoc
// @summary List application reports.
// @description List application reports.
// @description filters:
// @description - id
// @description - name
// @description - description
// @description - businessService
// @description - provider
// @description - name
// @description - version
// @description - sha
// @description - indirect
// @description - dep.provider
// @description - dep.name
// @description - dep.version
// @description - dep.sha
// @description - dep.indirect
// @description - dep.labels
// @description - application.id
// @description - application.name
// @description - businessService.id
// @description - businessService.name
// @description sort:
// @description - name
// @description - description
// @description - businessService
// @description - provider
// @description - name
// @description - version
// @description - sha
// @description - indirect
// @tags depappreports
// @produce json
// @success 200 {object} []api.DepAppReport
// @router /analyses/report/applications [get]
func (h AnalysisHandler) DepAppReports(ctx *gin.Context) {
	resources := []DepAppReport{}
	type M struct {
		ID              uint
		Name            string
		Description     string
		BusinessService string
		DepID           uint
		Provider        string
		DepName         string
		Version         string
		SHA             string
		Indirect        bool
		Labels          model.JSON
	}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "description", Kind: qf.STRING},
			{Field: "businessService", Kind: qf.STRING},
			{Field: "provider", Kind: qf.LITERAL},
			{Field: "name", Kind: qf.LITERAL},
			{Field: "version", Kind: qf.LITERAL},
			{Field: "sha", Kind: qf.LITERAL},
			{Field: "indirect", Kind: qf.LITERAL},
			{Field: "dep.provider", Kind: qf.LITERAL},
			{Field: "dep.name", Kind: qf.LITERAL},
			{Field: "dep.version", Kind: qf.LITERAL},
			{Field: "dep.sha", Kind: qf.LITERAL},
			{Field: "dep.indirect", Kind: qf.LITERAL},
			{Field: "dep.labels", Kind: qf.STRING, And: true},
			{Field: "application.id", Kind: qf.LITERAL},
			{Field: "application.name", Kind: qf.STRING},
			{Field: "businessService.id", Kind: qf.LITERAL},
			{Field: "businessService.name", Kind: qf.STRING},
			{Field: "tag.id", Kind: qf.LITERAL, And: true},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Sort
	sort := Sort{}
	err = sort.With(ctx, &M{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Inner Query
	q := h.DB(ctx)
	q = q.Select(
		"app.ID",
		"app.Name",
		"app.Description",
		"b.Name BusinessService",
		"d.ID DepID",
		"d.Provider",
		"d.Name DepName",
		"d.Version",
		"d.SHA",
		"d.Indirect",
		"d.Labels")
	q = q.Table("TechDependency d")
	q = q.Joins("LEFT JOIN Analysis a ON a.ID = d.AnalysisID")
	q = q.Joins("LEFT JOIN Application app ON app.ID = a.ApplicationID")
	q = q.Joins("LEFT OUTER JOIN BusinessService b ON b.ID = app.BusinessServiceID")
	q = q.Where("a.ID IN (?)", h.analysisIDs(ctx, filter))
	q = q.Where("d.ID IN (?)", h.depIDs(ctx, filter.Resource("dep")))
	// Find
	db := h.DB(ctx)
	db = db.Select("*")
	db = db.Table("(?)", q)
	db = filter.Where(db)
	db = sort.Sorted(db)
	var list []M
	var m M
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Render
	for i := range list {
		m := &list[i]
		r := DepAppReport{}
		r.ID = m.ID
		r.Name = m.Name
		r.Description = m.Description
		r.BusinessService = m.BusinessService
		r.Dependency.ID = m.DepID
		r.Dependency.Provider = m.Provider
		r.Dependency.Name = m.DepName
		r.Dependency.Version = m.Version
		r.Dependency.SHA = m.SHA
		r.Dependency.Indirect = m.Indirect
		if m.Labels != nil {
			_ = json.Unmarshal(m.Labels, &r.Dependency.Labels)
		}
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// appIDs provides application IDs.
// filter:
// - application.(id|name)
// - tag.id
func (h *AnalysisHandler) appIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Application{})
	q = q.Select("ID")
	appFilter := f.Resource("application")
	q = appFilter.Where(q)
	tagFilter := f.Resource("tag")
	if f, found := tagFilter.Field("id"); found {
		if f.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, f = range f.Expand() {
				f = f.As("TagID")
				iq := h.DB(ctx)
				iq = iq.Model(&model.ApplicationTag{})
				iq = iq.Select("applicationID ID")
				iq = f.Where(q)
				qs = append(qs, iq)
			}
			q = q.Where("ID IN (?)", model.Intersect(qs...))
		} else {
			f = f.As("TagID")
			iq := h.DB(ctx)
			iq = iq.Model(&model.ApplicationTag{})
			iq = iq.Select("ApplicationID ID")
			iq = f.Where(iq)
			q = q.Where("ID IN (?)", iq)
		}
	}
	bsFilter := f.Resource("businessService")
	if !bsFilter.Empty() {
		iq := h.DB(ctx)
		iq = iq.Model(&model.BusinessService{})
		iq = iq.Select("ID")
		iq = bsFilter.Where(iq)
		q = q.Where("BusinessServiceID IN (?)", iq)
		return
	}
	return
}

// analysisIDs provides analysis IDs.
func (h *AnalysisHandler) analysisIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Analysis{})
	q = q.Select("MAX(ID)")
	q = q.Where("ApplicationID IN (?)", h.appIDs(ctx, f))
	q = q.Group("ApplicationID")
	return
}

// issueIDs returns issue filtered issue IDs.
// Filter:
//
//	issue.*
func (h *AnalysisHandler) issueIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Issue{})
	q = q.Select("ID")
	q = f.Where(q, "-Labels")
	filter := f
	if f, found := filter.Field("labels"); found {
		if f.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, f = range f.Expand() {
				f = f.As("json_each.value")
				iq := h.DB(ctx)
				iq = iq.Table("Issue")
				iq = iq.Joins("m ,json_each(Labels)")
				iq = iq.Select("m.ID")
				iq = f.Where(iq)
				qs = append(qs, iq)
			}
			q = q.Where("ID IN (?)", model.Intersect(qs...))
		} else {
			f = f.As("json_each.value")
			iq := h.DB(ctx)
			iq = iq.Table("Issue")
			iq = iq.Joins("m ,json_each(Labels)")
			iq = iq.Select("m.ID")
			iq = f.Where(iq)
			q = q.Where("ID IN (?)", iq)
		}
	}
	return
}

// depIDs returns issue filtered issue IDs.
// Filter:
//
//	techDeps.*
func (h *AnalysisHandler) depIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.TechDependency{})
	q = q.Select("ID")
	q = f.Where(q, "-Labels")
	filter := f
	if f, found := filter.Field("labels"); found {
		if f.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, f = range f.Expand() {
				f = f.As("json_each.value")
				iq := h.DB(ctx)
				iq = iq.Table("TechDependency")
				iq = iq.Joins("m ,json_each(Labels)")
				iq = iq.Select("m.ID")
				iq = f.Where(iq)
				qs = append(qs, iq)
			}
			q = q.Where("ID IN (?)", model.Intersect(qs...))
		} else {
			f = f.As("json_each.value")
			iq := h.DB(ctx)
			iq = iq.Table("TechDependency")
			iq = iq.Joins("m ,json_each(Labels)")
			iq = iq.Select("m.ID")
			iq = f.Where(iq)
			q = q.Where("ID IN (?)", iq)
		}
	}
	return
}

// archive
// - Set the 'archived' flag.
// - Set the 'summary' field with archived issues.
// - Delete issues.
// - Delete dependencies.
func (h *AnalysisHandler) archive(ctx *gin.Context) (err error) {
	appId := h.pk(ctx)
	var unarchived []model.Analysis
	db := h.DB(ctx)
	db = db.Where("ApplicationID", appId)
	db = db.Where("Archived", false)
	err = db.Find(&unarchived).Error
	if err != nil {
		return
	}
	for _, m := range unarchived {
		db := h.DB(ctx)
		db = db.Select(
			"i.RuleSet",
			"i.Rule",
			"i.Name",
			"i.Description",
			"i.Category",
			"i.Effort",
			"COUNT(n.ID) Incidents")
		db = db.Table("Issue i,")
		db = db.Joins("Incident n")
		db = db.Where("n.IssueID = i.ID")
		db = db.Where("i.AnalysisID", m.ID)
		db = db.Group("i.ID")
		summary := []ArchivedIssue{}
		err = db.Scan(&summary).Error
		if err != nil {
			return
		}
		db = h.DB(ctx)
		db = db.Model(m)
		db = db.Omit(clause.Associations)
		m.Archived = true
		m.Summary, _ = json.Marshal(summary)
		err = db.Updates(h.fields(&m)).Error
		if err != nil {
			return
		}
		db = h.DB(ctx)
		db = db.Where("AnalysisID", m.ID)
		err = db.Delete(&model.Issue{}).Error
		if err != nil {
			return
		}
		db = h.DB(ctx)
		db = db.Where("AnalysisID", m.ID)
		err = db.Delete(&model.TechDependency{}).Error
		if err != nil {
			return
		}
	}

	return
}

// Analysis REST resource.
type Analysis struct {
	Resource     `yaml:",inline"`
	Effort       int              `json:"effort"`
	Archived     bool             `json:"archived,omitempty" yaml:",omitempty"`
	Issues       []Issue          `json:"issues,omitempty" yaml:",omitempty"`
	Dependencies []TechDependency `json:"dependencies,omitempty" yaml:",omitempty"`
	Summary      []ArchivedIssue  `json:"summary,omitempty" yaml:",omitempty" swaggertype:"object"`
}

// With updates the resource with the model.
func (r *Analysis) With(m *model.Analysis) {
	r.Resource.With(&m.Model)
	r.Effort = m.Effort
	r.Archived = m.Archived
	r.Issues = []Issue{}
	for i := range m.Issues {
		n := Issue{}
		n.With(&m.Issues[i])
		r.Issues = append(
			r.Issues,
			n)
	}
	r.Dependencies = []TechDependency{}
	for i := range m.Dependencies {
		n := TechDependency{}
		n.With(&m.Dependencies[i])
		r.Dependencies = append(
			r.Dependencies,
			n)
	}
	if m.Summary != nil {
		_ = json.Unmarshal(m.Summary, &r.Summary)
	}
}

// Model builds a model.
func (r *Analysis) Model() (m *model.Analysis) {
	m = &model.Analysis{}
	m.Effort = r.Effort
	m.Issues = []model.Issue{}
	for i := range r.Issues {
		n := r.Issues[i].Model()
		m.Issues = append(
			m.Issues,
			*n)
	}
	m.Dependencies = []model.TechDependency{}
	for i := range r.Dependencies {
		n := r.Dependencies[i].Model()
		m.Dependencies = append(
			m.Dependencies,
			*n)
	}
	return
}

// Issue REST resource.
type Issue struct {
	Resource    `yaml:",inline"`
	RuleSet     string     `json:"ruleset" binding:"required"`
	Rule        string     `json:"rule" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty" yaml:",omitempty"`
	Category    string     `json:"category" binding:"required"`
	Effort      int        `json:"effort,omitempty" yaml:",omitempty"`
	Incidents   []Incident `json:"incidents,omitempty" yaml:",omitempty"`
	Links       []Link     `json:"links,omitempty" yaml:",omitempty"`
	Facts       FactMap    `json:"facts,omitempty" yaml:",omitempty"`
	Labels      []string   `json:"labels"`
}

// With updates the resource with the model.
func (r *Issue) With(m *model.Issue) {
	r.Resource.With(&m.Model)
	r.RuleSet = m.RuleSet
	r.Rule = m.Rule
	r.Name = m.Name
	r.Description = m.Description
	r.Category = m.Category
	r.Incidents = []Incident{}
	for i := range m.Incidents {
		n := Incident{}
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

// Model builds a model.
func (r *Issue) Model() (m *model.Issue) {
	m = &model.Issue{}
	m.RuleSet = r.RuleSet
	m.Rule = r.Rule
	m.Name = r.Name
	m.Description = r.Description
	m.Category = r.Category
	m.Incidents = []model.Incident{}
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

// TechDependency REST resource.
type TechDependency struct {
	Resource `yaml:",inline"`
	Provider string   `json:"provider" yaml:",omitempty"`
	Name     string   `json:"name" binding:"required"`
	Version  string   `json:"version,omitempty" yaml:",omitempty"`
	Indirect bool     `json:"indirect,omitempty" yaml:",omitempty"`
	Labels   []string `json:"labels,omitempty" yaml:",omitempty"`
	SHA      string   `json:"sha,omitempty" yaml:",omitempty"`
}

// With updates the resource with the model.
func (r *TechDependency) With(m *model.TechDependency) {
	r.Resource.With(&m.Model)
	r.Provider = m.Provider
	r.Name = m.Name
	r.Version = m.Version
	r.Indirect = m.Indirect
	r.SHA = m.SHA
	if m.Labels != nil {
		_ = json.Unmarshal(m.Labels, &r.Labels)
	}
}

// Model builds a model.
func (r *TechDependency) Model() (m *model.TechDependency) {
	sort.Strings(r.Labels)
	m = &model.TechDependency{}
	m.Name = r.Name
	m.Version = r.Version
	m.Provider = r.Provider
	m.Indirect = r.Indirect
	m.Labels, _ = json.Marshal(r.Labels)
	m.SHA = r.SHA
	return
}

// Incident REST resource.
type Incident struct {
	Resource `yaml:",inline"`
	File     string  `json:"file"`
	Line     int     `json:"line"`
	Message  string  `json:"message"`
	CodeSnip string  `json:"codeSnip" yaml:"codeSnip"`
	Facts    FactMap `json:"facts"`
}

// With updates the resource with the model.
func (r *Incident) With(m *model.Incident) {
	r.Resource.With(&m.Model)
	r.File = m.File
	r.Line = m.Line
	r.Message = m.Message
	r.CodeSnip = m.CodeSnip
	if m.Facts != nil {
		_ = json.Unmarshal(m.Facts, &r.Facts)
	}
}

// Model builds a model.
func (r *Incident) Model() (m *model.Incident) {
	m = &model.Incident{}
	m.File = r.File
	m.Line = r.Line
	m.Message = r.Message
	m.CodeSnip = r.CodeSnip
	m.Facts, _ = json.Marshal(r.Facts)
	return
}

// Link analysis report link.
type Link struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty" yaml:",omitempty"`
}

// ArchivedIssue created when issues are archived.
type ArchivedIssue model.ArchivedIssue

// RuleReport REST resource.
type RuleReport struct {
	RuleSet      string   `json:"ruleset"`
	Rule         string   `json:"rule"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Effort       int      `json:"effort"`
	Labels       []string `json:"labels"`
	Links        []Link   `json:"links"`
	Applications int      `json:"applications"`
}

// IssueReport REST resource.
type IssueReport struct {
	ID          uint     `json:"id"`
	RuleSet     string   `json:"ruleset"`
	Rule        string   `json:"rule"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Effort      int      `json:"effort"`
	Labels      []string `json:"labels"`
	Links       []Link   `json:"links"`
	Files       int      `json:"files"`
}

// IssueAppReport REST resource.
type IssueAppReport struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	BusinessService string `json:"businessService"`
	Effort          int    `json:"effort"`
	Incidents       int    `json:"incidents"`
	Files           int    `json:"files"`
	Issue           struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		RuleSet     string `json:"ruleset"`
		Rule        string `json:"rule"`
	} `json:"issue"`
}

// FileReport REST resource.
type FileReport struct {
	IssueID   uint   `json:"issueId" yaml:"issueId"`
	File      string `json:"file"`
	Incidents int    `json:"incidents"`
	Effort    int    `json:"effort"`
}

// DepReport REST resource.
type DepReport struct {
	Provider     string   `json:"provider"`
	Name         string   `json:"name"`
	Labels       []string `json:"labels"`
	Applications int      `json:"applications"`
}

// DepAppReport REST resource.
type DepAppReport struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	BusinessService string `json:"businessService"`
	Dependency      struct {
		ID       uint     `json:"id"`
		Provider string   `json:"provider"`
		Name     string   `json:"name"`
		Version  string   `json:"version"`
		SHA      string   `json:"sha"`
		Indirect bool     `json:"indirect"`
		Labels   []string `json:"labels"`
	} `json:"dependency"`
}

// FactMap map.
type FactMap map[string]interface{}

// AnalysisWriter used to create a file containing an analysis.
type AnalysisWriter struct {
	encoder
	ctx *gin.Context
}

// db returns a db client.
func (r *AnalysisWriter) db() (db *gorm.DB) {
	rtx := WithContext(r.ctx)
	db = rtx.DB.Debug()
	return
}

// Create an analysis file and returns the path.
func (r *AnalysisWriter) Create(id uint) (path string, err error) {
	ext := ".json"
	accepted := r.ctx.NegotiateFormat(BindMIMEs...)
	switch accepted {
	case "",
		binding.MIMEPOSTForm,
		binding.MIMEJSON:
	case binding.MIMEYAML:
		ext = ".yaml"
	default:
		err = &BadRequestError{"MIME not supported."}
	}
	file, err := os.CreateTemp("", "report-*"+ext)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()
	path = file.Name()
	err = r.Write(id, file)
	return
}

// Write the analysis file.
func (r *AnalysisWriter) Write(id uint, output io.Writer) (err error) {
	m := &model.Analysis{}
	db := r.db()
	err = db.First(m, id).Error
	if err != nil {
		return
	}
	r.encoder, err = r.newEncoder(output)
	if err != nil {
		return
	}
	r.begin()
	rx := &Analysis{}
	rx.With(m)
	r.embed(rx)
	err = r.addIssues(m)
	if err != nil {
		return
	}
	err = r.addDeps(m)
	if err != nil {
		return
	}
	r.end()
	return
}

// newEncoder returns an encoder.
func (r *AnalysisWriter) newEncoder(output io.Writer) (encoder encoder, err error) {
	accepted := r.ctx.NegotiateFormat(BindMIMEs...)
	switch accepted {
	case "",
		binding.MIMEPOSTForm,
		binding.MIMEJSON:
		encoder = &jsonEncoder{output: output}
	case binding.MIMEYAML:
		encoder = &yamlEncoder{output: output}
	default:
		err = &BadRequestError{"MIME not supported."}
	}

	return
}

// addIssues writes issues.
func (r *AnalysisWriter) addIssues(m *model.Analysis) (err error) {
	r.field("issues")
	r.beginList()
	batch := 10
	for b := 0; ; b += batch {
		db := r.db()
		db = db.Preload("Incidents")
		db = db.Limit(batch)
		db = db.Offset(b)
		var issues []model.Issue
		err = db.Find(&issues, "AnalysisID", m.ID).Error
		if err != nil {
			return
		}
		if len(issues) == 0 {
			break
		}
		for i := range issues {
			issue := Issue{}
			issue.With(&issues[i])
			r.writeItem(b, i, issue)
		}
	}
	r.endList()
	return
}

// addDeps writes dependencies.
func (r *AnalysisWriter) addDeps(m *model.Analysis) (err error) {
	r.field("dependencies")
	r.beginList()
	batch := 100
	for b := 0; ; b += batch {
		db := r.db()
		db = db.Limit(batch)
		db = db.Offset(b)
		var deps []model.TechDependency
		err = db.Find(&deps, "AnalysisID", m.ID).Error
		if err != nil {
			return
		}
		if len(deps) == 0 {
			break
		}
		for i := range deps {
			d := TechDependency{}
			d.With(&deps[i])
			r.writeItem(b, i, d)
		}
	}
	r.endList()
	return
}

// ReportWriter analysis report writer.
type ReportWriter struct {
	encoder
	ctx *gin.Context
}

// db returns a db client.
func (r *ReportWriter) db() (db *gorm.DB) {
	rtx := WithContext(r.ctx)
	db = rtx.DB.Debug()
	return
}

// Write builds and streams the analysis report.
func (r *ReportWriter) Write(id uint) {
	reportDir := Settings.Analysis.ReportPath
	path, err := r.buildOutput(id)
	if err != nil {
		_ = r.ctx.Error(err)
		return
	}
	defer func() {
		_ = os.Remove(path)
	}()
	tarWriter := tar.NewWriter(r.ctx.Writer)
	defer func() {
		tarWriter.Close()
	}()
	filter := tar.NewFilter(reportDir)
	filter.Exclude("output.js")
	tarWriter.Filter = filter
	err = tarWriter.AssertDir(Settings.Analysis.ReportPath)
	if err != nil {
		_ = r.ctx.Error(err)
		return
	}
	err = tarWriter.AssertFile(path)
	if err != nil {
		_ = r.ctx.Error(err)
		return
	}
	r.ctx.Status(http.StatusOK)
	_ = tarWriter.AddDir(Settings.Analysis.ReportPath)
	_ = tarWriter.AddFile(path, "output.js")
	return
}

// buildOutput creates the report output.js file.
func (r *ReportWriter) buildOutput(id uint) (path string, err error) {
	m := &model.Analysis{}
	db := r.db()
	db = db.Preload("Application")
	db = db.Preload("Application.Tags")
	db = db.Preload("Application.Tags.Category")
	err = db.First(m, id).Error
	if err != nil {
		return
	}
	file, err := os.CreateTemp("", "output-*.js")
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()
	path = file.Name()
	r.encoder = &jsonEncoder{output: file}
	r.write("window[\"apps\"]=[")
	r.begin()
	r.field("id").writeStr(strconv.Itoa(int(m.Application.ID)))
	r.field("name").writeStr(m.Application.Name)
	r.field("analysis").writeStr(strconv.Itoa(int(m.ID)))
	aWriter := AnalysisWriter{ctx: r.ctx}
	aWriter.encoder = r.encoder
	err = aWriter.addIssues(m)
	if err != nil {
		return
	}
	err = aWriter.addDeps(m)
	if err != nil {
		return
	}
	err = r.addTags(m)
	if err != nil {
		return
	}
	r.end()
	r.write("]")
	return
}

// addTags writes tags.
func (r *ReportWriter) addTags(m *model.Analysis) (err error) {
	r.field("tags")
	r.beginList()
	for i := range m.Application.Tags {
		m := m.Application.Tags[i]
		tag := Tag{}
		tag.ID = m.ID
		tag.Name = m.Name
		tag.Category = Ref{
			ID:   m.Category.ID,
			Name: m.Category.Name,
		}
		r.writeItem(0, i, tag)
	}
	r.endList()
	return
}

type encoder interface {
	begin() encoder
	end() encoder
	write(s string) encoder
	writeStr(s string) encoder
	field(name string) encoder
	beginList() encoder
	endList() encoder
	writeItem(batch, index int, object any) encoder
	encode(object any) encoder
	embed(object any) encoder
}

type jsonEncoder struct {
	output io.Writer
	fields int
}

func (r *jsonEncoder) begin() encoder {
	r.write("{")
	return r
}

func (r *jsonEncoder) end() encoder {
	r.write("}")
	return r
}

func (r *jsonEncoder) write(s string) encoder {
	_, _ = r.output.Write([]byte(s))
	return r
}

func (r *jsonEncoder) writeStr(s string) encoder {
	r.write("\"" + s + "\"")
	return r
}

func (r *jsonEncoder) field(s string) encoder {
	if r.fields > 0 {
		r.write(",")
	}
	r.writeStr(s).write(":")
	r.fields++
	return r
}

func (r *jsonEncoder) beginList() encoder {
	r.write("[")
	return r
}

func (r *jsonEncoder) endList() encoder {
	r.write("]")
	return r
}

func (r *jsonEncoder) writeItem(batch, index int, object any) encoder {
	if batch > 0 || index > 0 {
		r.write(",")
	}
	r.encode(object)
	return r
}

func (r *jsonEncoder) encode(object any) encoder {
	encoder := json.NewEncoder(r.output)
	_ = encoder.Encode(object)
	return r
}

func (r *jsonEncoder) embed(object any) encoder {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	_ = encoder.Encode(object)
	s := b.String()
	mp := make(map[string]any)
	err := json.Unmarshal([]byte(s), &mp)
	if err == nil {
		r.fields += len(mp)
		s = s[1 : len(s)-2]
	}
	r.write(s)
	return r
}

type yamlEncoder struct {
	output io.Writer
	fields int
	depth  int
}

func (r *yamlEncoder) begin() encoder {
	r.write("---\n")
	return r
}

func (r *yamlEncoder) end() encoder {
	return r
}

func (r *yamlEncoder) write(s string) encoder {
	s += strings.Repeat("  ", r.depth)
	_, _ = r.output.Write([]byte(s))
	return r
}

func (r *yamlEncoder) writeStr(s string) encoder {
	r.write("\"" + s + "\"")
	return r
}

func (r *yamlEncoder) field(s string) encoder {
	if r.fields > 0 {
		r.write("\n")
	}
	r.write(s).write(": ")
	r.fields++
	return r
}

func (r *yamlEncoder) beginList() encoder {
	r.write("\n")
	r.depth++
	return r
}

func (r *yamlEncoder) endList() encoder {
	r.depth--
	return r
}

func (r *yamlEncoder) writeItem(batch, index int, object any) encoder {
	r.encode([]any{object})
	return r
}

func (r *yamlEncoder) encode(object any) encoder {
	encoder := yaml.NewEncoder(r.output)
	_ = encoder.Encode(object)
	return r
}

func (r *yamlEncoder) embed(object any) encoder {
	b := new(bytes.Buffer)
	encoder := yaml.NewEncoder(b)
	_ = encoder.Encode(object)
	s := b.String()
	mp := make(map[string]any)
	err := yaml.Unmarshal([]byte(s), &mp)
	if err == nil {
		r.fields += len(mp)
	}
	r.write(s)
	return r
}
