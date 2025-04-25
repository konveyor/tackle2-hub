package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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
	AnalysisArchiveRoot   = AnalysisRoot + "/archive"
	AnalysisIssuesRoot    = AnalysisRoot + "/issues"
	AnalysisIncidentsRoot = AnalysesIssueRoot + "/incidents"
	AnalysesDepsRoot      = AnalysesRoot + "/dependencies"
	AnalysesIssuesRoot    = AnalysesRoot + "/issues"
	AnalysesIssueRoot     = AnalysesIssuesRoot + "/:" + ID
	AnalysesIncidentsRoot = AnalysesRoot + "/incidents"
	AnalysesIncidentRoot  = AnalysesIncidentsRoot + "/:" + ID
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

// Manifest markers.
// The GS=\x1D (group separator).
const (
	BeginMainMarker   = "\x1DBEGIN-MAIN\x1D"
	EndMainMarker     = "\x1DEND-MAIN\x1D"
	BeginIssuesMarker = "\x1DBEGIN-ISSUES\x1D"
	EndIssuesMarker   = "\x1DEND-ISSUES\x1D"
	BeginDepsMarker   = "\x1DBEGIN-DEPS\x1D"
	EndDepsMarker     = "\x1DEND-DEPS\x1D"
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
	routeGroup.POST(AnalysisArchiveRoot, h.Archive)
	routeGroup.GET(AnalysesRoot, h.List)
	routeGroup.DELETE(AnalysisRoot, h.Delete)
	routeGroup.GET(AnalysesDepsRoot, h.Deps)
	routeGroup.GET(AnalysesIssuesRoot, h.Issues)
	routeGroup.GET(AnalysesIssueRoot, h.Issue)
	routeGroup.GET(AnalysesIncidentsRoot, h.Incidents)
	routeGroup.GET(AnalysesIncidentRoot, h.Incident)
	routeGroup.GET(AnalysisIssuesRoot, h.AnalysisIssues)
	routeGroup.GET(AnalysisIncidentsRoot, h.IssueIncidents)
	// Report
	routeGroup.GET(AnalysisReportRuleRoot, h.RuleReports)
	routeGroup.GET(AnalysisReportAppsIssuesRoot, h.AppIssueReports)
	routeGroup.GET(AnalysisReportIssuesAppsRoot, h.IssueAppReports)
	routeGroup.GET(AnalysisReportFileRoot, h.FileReports)
	routeGroup.GET(AnalysisReportDepsRoot, h.DepReports)
	routeGroup.GET(AnalysisReportDepsAppsRoot, h.DepAppReports)
	// Application
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.analyses"))
	routeGroup.POST(AppAnalysesRoot, Transaction, h.AppCreate)
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

// List godoc
// @summary List analyses.
// @description List analyses.
// @description Resources do not include relations.
// @tags analyses
// @produce json
// @success 200 {object} []api.Analysis
// @router /analyses [get]
func (h AnalysisHandler) List(ctx *gin.Context) {
	resources := []Analysis{}
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("id", "analysis\\.id")
	// sort.
	sort := Sort{}
	err = sort.With(ctx, &model.Analysis{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	db := h.DB(ctx)
	db = db.Model(&model.Analysis{})
	db = db.Joins("Application")
	db = db.Omit("Summary")
	db = filter.Where(db)
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
	// Render
	for i := range list {
		m := &list[i]
		r := Analysis{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
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

// Archive godoc
// @summary Archive an analysis (report) by ID.
// @description Archive an analysis (report) by ID.
// @tags analyses
// @produce octet-stream
// @success 204
// @router /analyses/{id}/archive [post]
// @param id path int true "Analysis ID"
func (h AnalysisHandler) Archive(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Analysis{}
	db := h.DB(ctx).Select(ID)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.archiveById(ctx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
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
	db = db.Joins("Application")
	db = db.Omit("Summary")
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
	// Render
	for i := range list {
		m := &list[i]
		r := Analysis{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AppCreate godoc
// @summary Create an analysis.
// @description Create an analysis.
// @description Form fields:
// @description file: A manifest file that contains 3 sections
// @description containing documents delimited by markers.
// @description The manifest must contain ALL markers even when sections are empty.
// @description Note: `^]` = `\x1D` = GS (group separator).
// @description Section markers:
// @description	  ^]BEGIN-MAIN^]
// @description	  ^]END-MAIN^]
// @description	  ^]BEGIN-ISSUES^]
// @description	  ^]END-ISSUES^]
// @description	  ^]BEGIN-DEPS^]
// @description	  ^]END-DEPS^]
// @description The encoding must be:
// @description - application/json
// @description - application/x-yaml
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
	if Settings.Analysis.ArchiverEnabled {
		err := h.archiveByApp(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}
	db := h.DB(ctx)
	//
	// Manifest
	fh := FileHandler{}
	name := fmt.Sprintf("app.%d.manifest", id)
	file, err := fh.create(ctx, name)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		err = fh.delete(ctx, file)
		if err != nil {
			_ = ctx.Error(err)
		}
	}()
	reader := &ManifestReader{}
	f, err := reader.open(file.Path, BeginMainMarker, EndMainMarker)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	d, err := h.Decoder(ctx, file.Encoding, reader)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	r := &Analysis{}
	err = d.Decode(r)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	analysis := r.Model()
	analysis.ApplicationID = id
	analysis.CreateUser = h.BaseHandler.CurrentUser(ctx)
	db.Logger = db.Logger.LogMode(logger.Error)
	err = db.Create(analysis).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	//
	// Issues
	reader = &ManifestReader{}
	f, err = reader.open(file.Path, BeginIssuesMarker, EndIssuesMarker)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	d, err = h.Decoder(ctx, file.Encoding, reader)
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
	reader = &ManifestReader{}
	f, err = reader.open(file.Path, BeginDepsMarker, EndDepsMarker)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = f.Close()
	}()
	d, err = h.Decoder(ctx, file.Encoding, reader)
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
	db = db.Preload("Application")
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
	// Render
	writer := IssueWriter{ctx: ctx}
	path, count, err := writer.Create(analysis.ID, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = os.Remove(path)
	}()
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusOK)
	ctx.File(path)
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
	// Render
	writer := IssueWriter{ctx: ctx}
	path, count, err := writer.Create(0, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = os.Remove(path)
	}()
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusOK)
	ctx.File(path)
}

// AnalysisIssues godoc
// @summary List issues for an analysis.
// @description List issues for an analysis.
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
// @router /analyses/{id}/issues [get]
// @param id path int true "Analysis ID"
func (h AnalysisHandler) AnalysisIssues(ctx *gin.Context) {
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
	// Render
	id := h.pk(ctx)
	writer := IssueWriter{ctx: ctx}
	path, count, err := writer.Create(id, filter)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = os.Remove(path)
	}()
	err = h.WithCount(ctx, count)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusOK)
	ctx.File(path)
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
// @summary List all incidents.
// @description List all incidents.
// @description filters:
// @description - file
// @description - issue.id
// @tags incidents
// @produce json
// @success 200 {object} []api.Incident
// @router /analyses/incidents [get]
func (h AnalysisHandler) Incidents(ctx *gin.Context) {
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "file", Kind: qf.STRING},
			{Field: "issue.id", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("issue.id", "issueid")
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

// IssueIncidents godoc
// @summary List incidents for an issue.
// @description List incidents for an issue.
// @description filters:
// @description - file
// @tags incidents
// @produce json
// @success 200 {object} []api.Incident
// @router /analyses/issues/{id}/incidents [get]
// @param id path int true "Issue ID"
func (h AnalysisHandler) IssueIncidents(ctx *gin.Context) {
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

// Incident godoc
// @summary Get an incident.
// @description Get an incident.
// @tags issue
// @produce json
// @success 200 {object} api.Incident
// @router /analyses/incidents/{id} [get]
// @param id path int true "Issue ID"
func (h AnalysisHandler) Incident(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Incident{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Incident{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
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
		r.Labels = m.Labels
		for _, l := range m.Links {
			r.Links = append(r.Links, Link(l))
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
		Files   int
		Message string
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
		"n.Message",
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
			Files: m.Files,
			// Append Incident Message to Description to provide more information on Issue detail
			// (workaround until Analyzer output Violation struct gets updated to provide better structured data)
			Description: fmt.Sprintf("%s\n\n%s", m.Description, m.Message),
			Category:    m.Category,
			RuleSet:     m.RuleSet,
			Rule:        m.Rule,
			Name:        m.Name,
			ID:          m.ID,
		}
		resources = append(resources, r)
		r.Labels = m.Labels
		for _, l := range m.Links {
			r.Links = append(r.Links, Link(l))
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
		"i.Effort * COUNT(n.ID) as Effort",
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
	db = db.Where("AnalysisID IN (?)", h.analysesIDs(ctx, filter))
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

// AnalysisDeps godoc
// @summary List analysis dependencies.
// @description List analysis dependencies.
// @description filters:
// @description - name
// @description - version
// @description - sha
// @description - indirect
// @description - labels
// @tags dependencies
// @produce json
// @success 200 {object} []api.TechDependency
// @router /analyses/{id}/dependencies [get]
// @param id path int true "Analysis ID"
func (h AnalysisHandler) AnalysisDeps(ctx *gin.Context) {
	resources := []TechDependency{}
	// Filter
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
	db = db.Where("AnalysisID = ?", h.pk(ctx))
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
		"d.Labels",
		"COUNT(distinct d.AnalysisID) Applications")
	q = q.Table("TechDependency d")
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
		for _, s := range m.Labels {
			if s != "" {
				r.Labels = append(r.Labels, s)
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
func (h *AnalysisHandler) analysesIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Analysis{})
	q = q.Select("ID")
	q = q.Where("ApplicationID IN (?)", h.appIDs(ctx, f))
	q = q.Group("ApplicationID")
	return
}

// analysisIDs provides LATEST analysis IDs.
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

// archiveById
// - Set the 'archived' flag.
// - Set the 'summary' field with archived issues.
// - Delete issues.
// - Delete dependencies.
func (h *AnalysisHandler) archiveById(ctx *gin.Context) (err error) {
	id := h.pk(ctx)
	db := h.DB(ctx)
	db = db.Where("id", id)
	err = h.archive(ctx, db)
	return
}

// archiveByApp
// - Set the 'archived' flag.
// - Set the 'summary' field with archived issues.
// - Delete issues.
// - Delete dependencies.
func (h *AnalysisHandler) archiveByApp(ctx *gin.Context) (err error) {
	id := h.pk(ctx)
	db := h.DB(ctx)
	db = db.Where("ApplicationID", id)
	err = h.archive(ctx, db)
	return
}

// archive
// - Set the 'archived' flag.
// - Set the 'summary' field with archived issues.
// - Delete issues.
// - Delete dependencies.
func (h *AnalysisHandler) archive(ctx *gin.Context, q *gorm.DB) (err error) {
	var unarchived []model.Analysis
	q = q.Where("Archived", false)
	err = q.Find(&unarchived).Error
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
		summary := []model.ArchivedIssue{}
		err = db.Scan(&summary).Error
		if err != nil {
			return
		}
		db = h.DB(ctx)
		db = db.Model(m)
		db = db.Omit(clause.Associations)
		m.Archived = true
		m.Summary = summary
		err = db.Save(&m).Error
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
	Application  Ref              `json:"application"`
	Effort       int              `json:"effort"`
	Commit       string           `json:"commit,omitempty" yaml:",omitempty"`
	Archived     bool             `json:"archived,omitempty" yaml:",omitempty"`
	Issues       []Issue          `json:"issues,omitempty" yaml:",omitempty"`
	Dependencies []TechDependency `json:"dependencies,omitempty" yaml:",omitempty"`
	Summary      []ArchivedIssue  `json:"summary,omitempty" yaml:",omitempty" swaggertype:"object"`
}

// With updates the resource with the model.
func (r *Analysis) With(m *model.Analysis) {
	r.Resource.With(&m.Model)
	r.Application = r.ref(m.ApplicationID, m.Application)
	r.Effort = m.Effort
	r.Commit = m.Commit
	r.Archived = m.Archived
}

// Model builds a model.
func (r *Analysis) Model() (m *model.Analysis) {
	m = &model.Analysis{}
	m.Effort = r.Effort
	m.Commit = r.Commit
	m.Issues = []model.Issue{}
	return
}

// Issue REST resource.
type Issue struct {
	Resource    `yaml:",inline"`
	Analysis    uint       `json:"analysis"`
	RuleSet     string     `json:"ruleset" binding:"required"`
	Rule        string     `json:"rule" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty" yaml:",omitempty"`
	Category    string     `json:"category" binding:"required"`
	Effort      int        `json:"effort,omitempty" yaml:",omitempty"`
	Incidents   []Incident `json:"incidents,omitempty" yaml:",omitempty"`
	Links       []Link     `json:"links,omitempty" yaml:",omitempty"`
	Facts       Map        `json:"facts,omitempty" yaml:",omitempty"`
	Labels      []string   `json:"labels"`
}

// With updates the resource with the model.
func (r *Issue) With(m *model.Issue) {
	r.Resource.With(&m.Model)
	r.Analysis = m.AnalysisID
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
	for _, l := range m.Links {
		r.Links = append(r.Links, Link(l))
	}
	r.Facts = m.Facts
	r.Labels = m.Labels
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
	for _, l := range r.Links {
		m.Links = append(m.Links, model.Link(l))
	}
	m.Facts = r.Facts
	m.Labels = r.Labels
	m.Effort = r.Effort
	return
}

// TechDependency REST resource.
type TechDependency struct {
	Resource `yaml:",inline"`
	Analysis uint     `json:"analysis"`
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
	r.Analysis = m.AnalysisID
	r.Provider = m.Provider
	r.Name = m.Name
	r.Version = m.Version
	r.Indirect = m.Indirect
	r.SHA = m.SHA
	r.Labels = m.Labels
}

// Model builds a model.
func (r *TechDependency) Model() (m *model.TechDependency) {
	sort.Strings(r.Labels)
	m = &model.TechDependency{}
	m.Name = r.Name
	m.Version = r.Version
	m.Provider = r.Provider
	m.Indirect = r.Indirect
	m.Labels = r.Labels
	m.SHA = r.SHA
	return
}

// Incident REST resource.
type Incident struct {
	Resource `yaml:",inline"`
	Issue    uint   `json:"issue"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	CodeSnip string `json:"codeSnip" yaml:"codeSnip"`
	Facts    Map    `json:"facts"`
}

// With updates the resource with the model.
func (r *Incident) With(m *model.Incident) {
	r.Resource.With(&m.Model)
	r.Issue = m.IssueID
	r.File = m.File
	r.Line = m.Line
	r.Message = m.Message
	r.CodeSnip = m.CodeSnip
	r.Facts = m.Facts
}

// Model builds a model.
func (r *Incident) Model() (m *model.Incident) {
	m = &model.Incident{}
	m.File = r.File
	m.Line = r.Line
	m.Message = r.Message
	m.CodeSnip = r.CodeSnip
	m.Facts = r.Facts
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

// IssueWriter used to create a file containing issues.
type IssueWriter struct {
	encoder
	ctx *gin.Context
}

// Create an issues file and returns the path.
func (r *IssueWriter) Create(id uint, filter qf.Filter) (path string, count int64, err error) {
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
	file, err := os.CreateTemp("", "issue-*"+ext)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()
	path = file.Name()
	count, err = r.Write(id, filter, file)
	return
}

// db returns a db client.
func (r *IssueWriter) db() (db *gorm.DB) {
	rtx := RichContext(r.ctx)
	db = rtx.DB.Debug()
	return
}

// Write the analysis file.
func (r *IssueWriter) Write(id uint, filter qf.Filter, output io.Writer) (count int64, err error) {
	r.encoder, err = r.newEncoder(output)
	if err != nil {
		return
	}
	page := Page{}
	page.With(r.ctx)
	sort := Sort{}
	err = sort.With(r.ctx, &model.Issue{})
	if err != nil {
		return
	}
	r.beginList()
	batch := 10
	for b := page.Offset; ; b += batch {
		db := r.db()
		if id > 0 {
			db = db.Where("AnalysisID", id)
		}
		db = filter.Where(db)
		db = db.Preload("Incidents")
		db = db.Limit(batch)
		db = db.Offset(b)
		db = sort.Sorted(db)
		var issues []model.Issue
		err = db.Find(&issues).Error
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
			count++
		}
	}
	r.endList()
	return
}

// newEncoder returns an encoder.
func (r *IssueWriter) newEncoder(output io.Writer) (encoder encoder, err error) {
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

// AnalysisWriter used to create a file containing an analysis.
type AnalysisWriter struct {
	encoder
	ctx *gin.Context
}

// db returns a db client.
func (r *AnalysisWriter) db() (db *gorm.DB) {
	rtx := RichContext(r.ctx)
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
	rtx := RichContext(r.ctx)
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

// ManifestReader analysis manifest reader.
// The manifest contains 3 sections containing documents delimited by markers.
// The manifest must contain ALL markers even when sections are empty.
// Note: `^]` = `\x1D` = GS (group separator).
// Section markers:
//
//	^]BEGIN-MAIN^]
//	^]END-MAIN^]
//	^]BEGIN-ISSUES^]
//	^]END-ISSUES^]
//	^]BEGIN-DEPS^]
//	^]END-DEPS^]
type ManifestReader struct {
	file   *os.File
	marker map[string]int64
	begin  int64
	end    int64
	read   int64
}

// scan manifest and catalog position of markers.
func (r *ManifestReader) scan(path string) (err error) {
	if r.marker != nil {
		return
	}
	r.file, err = os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = r.file.Close()
	}()
	pattern, err := regexp.Compile(`^\x1D[A-Z-]+\x1D$`)
	if err != nil {
		return
	}
	p := int64(0)
	r.marker = make(map[string]int64)
	scanner := bufio.NewScanner(r.file)
	for scanner.Scan() {
		content := scanner.Text()
		matched := strings.TrimSpace(content)
		if pattern.Match([]byte(matched)) {
			r.marker[matched] = p
		}
		p += int64(len(content))
		p++
	}

	return
}

// open returns a read delimited by the specified markers.
func (r *ManifestReader) open(path, begin, end string) (reader io.ReadCloser, err error) {
	found := false
	err = r.scan(path)
	if err != nil {
		return
	}
	r.begin, found = r.marker[begin]
	if !found {
		err = &BadRequestError{
			Reason: fmt.Sprintf("marker: %s not found.", begin),
		}
		return
	}
	r.end, found = r.marker[end]
	if !found {
		err = &BadRequestError{
			Reason: fmt.Sprintf("marker: %s not found.", end),
		}
		return
	}
	if r.begin >= r.end {
		err = &BadRequestError{
			Reason: fmt.Sprintf("marker: %s must preceed %s.", begin, end),
		}
		return
	}
	r.begin += int64(len(begin))
	r.begin++
	r.read = r.end - r.begin
	r.file, err = os.Open(path)
	if err != nil {
		return
	}
	_, err = r.file.Seek(r.begin, io.SeekStart)
	reader = r
	return
}

// Read bytes.
func (r *ManifestReader) Read(b []byte) (n int, err error) {
	n, err = r.file.Read(b)
	if n == 0 || err != nil {
		return
	}
	if int64(n) > r.read {
		n = int(r.read)
	}
	r.read -= int64(n)
	if n < 1 {
		err = io.EOF
	}
	return
}

// Close the reader.
func (r *ManifestReader) Close() (err error) {
	err = r.file.Close()
	return
}
