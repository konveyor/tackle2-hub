package api

import (
	"bufio"
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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Routes
const (
	AnalysesRoot          = "/analyses"
	AnalysisRoot          = AnalysesRoot + "/:" + ID
	AnalysisArchiveRoot   = AnalysisRoot + "/archive"
	AnalysisInsightsRoot  = AnalysisRoot + "/insights"
	AnalysisIncidentsRoot = AnalysesInsightRoot + "/incidents"
	AnalysesDepsRoot      = AnalysesRoot + "/dependencies"
	AnalysesInsightsRoot  = AnalysesRoot + "/insights"
	AnalysesInsightRoot   = AnalysesInsightsRoot + "/:" + ID
	AnalysesIncidentsRoot = AnalysesRoot + "/incidents"
	AnalysesIncidentRoot  = AnalysesIncidentsRoot + "/:" + ID
	//
	AnalysesReportRoot             = AnalysesRoot + "/report"
	AnalysisReportDepsRoot         = AnalysesReportRoot + "/dependencies"
	AnalysisReportRuleRoot         = AnalysesReportRoot + "/rules"
	AnalysisReportInsightsRoot     = AnalysesReportRoot + "/insights"
	AnalysisReportAppsRoot         = AnalysesReportRoot + "/applications"
	AnalysisReportInsightRoot      = AnalysisReportInsightsRoot + "/:" + ID
	AnalysisReportInsightsAppsRoot = AnalysisReportInsightsRoot + "/applications"
	AnalysisReportDepsAppsRoot     = AnalysisReportDepsRoot + "/applications"
	AnalysisReportAppsInsightsRoot = AnalysisReportAppsRoot + "/:" + ID + "/insights"
	AnalysisReportFileRoot         = AnalysisReportInsightRoot + "/files"
	//
	AppAnalysesRoot         = ApplicationRoot + "/analyses"
	AppAnalysisRoot         = ApplicationRoot + "/analysis"
	AppAnalysisReportRoot   = AppAnalysisRoot + "/report"
	AppAnalysisDepsRoot     = AppAnalysisRoot + "/dependencies"
	AppAnalysisInsightsRoot = AppAnalysisRoot + "/insights"
)

// Manifest markers.
// The GS=\x1D (group separator).
const (
	BeginMainMarker     = "\x1DBEGIN-MAIN\x1D"
	EndMainMarker       = "\x1DEND-MAIN\x1D"
	BeginInsightsMarker = "\x1DBEGIN-INSIGHTS\x1D"
	EndInsightsMarker   = "\x1DEND-INSIGHTS\x1D"
	BeginDepsMarker     = "\x1DBEGIN-DEPS\x1D"
	EndDepsMarker       = "\x1DEND-DEPS\x1D"
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
	routeGroup.GET(AnalysesInsightsRoot, h.Insights)
	routeGroup.GET(AnalysesInsightRoot, h.Insight)
	routeGroup.GET(AnalysesIncidentsRoot, h.Incidents)
	routeGroup.GET(AnalysesIncidentRoot, h.Incident)
	routeGroup.GET(AnalysisInsightsRoot, h.AnalysisInsights)
	routeGroup.GET(AnalysisIncidentsRoot, h.InsightIncidents)
	// Report
	routeGroup.GET(AnalysisReportRuleRoot, h.RuleReports)
	routeGroup.GET(AnalysisReportAppsInsightsRoot, h.AppInsightReports)
	routeGroup.GET(AnalysisReportInsightsAppsRoot, h.InsightAppReports)
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
	routeGroup.GET(AppAnalysisInsightsRoot, h.AppInsights)
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
	err := sort.With(ctx, &model.Insight{})
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
// @description	  ^]BEGIN-INSIGHTS^]
// @description	  ^]END-INSIGHTS^]
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
	opened := []io.ReadCloser{}
	defer func() {
		for _, r := range opened {
			_ = r.Close()
		}
	}()
	//
	// Main
	reader := &ManifestReader{}
	err = reader.Open(file.Path, BeginMainMarker, EndMainMarker)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	} else {
		opened = append(opened, reader)
	}
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
	// Insights
	reader = &ManifestReader{}
	err = reader.Open(file.Path, BeginInsightsMarker, EndInsightsMarker)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	} else {
		opened = append(opened, reader)
	}
	d, err = h.Decoder(ctx, file.Encoding, reader)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	}
	for {
		r := &Insight{}
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
	err = reader.Open(file.Path, BeginDepsMarker, EndDepsMarker)
	if err != nil {
		err = &BadRequestError{err.Error()}
		_ = ctx.Error(err)
		return
	} else {
		opened = append(opened, reader)
	}
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

// AppInsights godoc
// @summary List application insights.
// @description List application insights.
// @description filters:
// @description - ruleset
// @description - rule
// @description - name
// @description - category
// @description - effort
// @description - labels
// @tags insights
// @produce json
// @success 200 {object} []api.Insight
// @router /application/{id}/analysis/insights [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppInsights(ctx *gin.Context) {
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
	writer := InsightWriter{ctx: ctx}
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

// Insights godoc
// @summary List all insights.
// @description List all insights.
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
// @tags insights
// @produce json
// @success 200 {object} []api.Insight
// @router /analyses/insights [get]
func (h AnalysisHandler) Insights(ctx *gin.Context) {
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
	writer := InsightWriter{ctx: ctx}
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

// AnalysisInsights godoc
// @summary List insights for an analysis.
// @description List insights for an analysis.
// @description filters:
// @description - ruleset
// @description - rule
// @description - name
// @description - category
// @description - effort
// @description - labels
// @tags insights
// @produce json
// @success 200 {object} []api.Insight
// @router /analyses/{id}/insights [get]
// @param id path int true "Analysis ID"
func (h AnalysisHandler) AnalysisInsights(ctx *gin.Context) {
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
	writer := InsightWriter{ctx: ctx}
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

// Insight godoc
// @summary Get an insight.
// @description Get an insight.
// @tags insight
// @produce json
// @success 200 {object} api.Insight
// @router /analyses/insights/{id} [get]
// @param id path int true "Insight ID"
func (h AnalysisHandler) Insight(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Insight{}
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Insight{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// Incidents godoc
// @summary List all incidents.
// @description List all incidents.
// @description filters:
// @description - file
// @description - insight.id
// @tags incidents
// @produce json
// @success 200 {object} []api.Incident
// @router /analyses/incidents [get]
func (h AnalysisHandler) Incidents(ctx *gin.Context) {
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "file", Kind: qf.STRING},
			{Field: "insight.id", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("insight.id", "insightid")
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

// InsightIncidents godoc
// @summary List incidents for an insight.
// @description List incidents for an insight.
// @description filters:
// @description - file
// @tags incidents
// @produce json
// @success 200 {object} []api.Incident
// @router /analyses/insights/{id}/incidents [get]
// @param id path int true "Insight ID"
func (h AnalysisHandler) InsightIncidents(ctx *gin.Context) {
	insightId := ctx.Param(ID)
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
	db = db.Where("InsightID", insightId)
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
// @tags incident
// @produce json
// @success 200 {object} api.Incident
// @router /analyses/incidents/{id} [get]
// @param id path int true "Insight ID"
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
// @description Each report collates insights by ruleset/rule.
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
		model.Insight
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
	q = q.Table("Insight i,")
	q = q.Joins("Analysis a")
	q = q.Where("a.ID = i.AnalysisID")
	q = q.Where("a.ID in (?)", h.analysisIDs(ctx, filter))
	q = q.Where("i.ID IN (?)", h.insightIDs(ctx, filter))
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

// AppInsightReports godoc
// @summary List application insight reports.
// @description Each report collates insights by ruleset/rule.
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
// @tags insightreport
// @produce json
// @success 200 {object} []api.InsightReport
// @router /analyses/report/applications/{id}/insights [get]
// @param id path int true "Application ID"
func (h AnalysisHandler) AppInsightReports(ctx *gin.Context) {
	resources := []*InsightReport{}
	type M struct {
		model.Insight
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
	q = q.Table("Insight i,")
	q = q.Joins("Incident n")
	q = q.Where("i.ID = n.InsightID")
	q = q.Where("i.ID IN (?)", h.insightIDs(ctx, filter))
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
		r := &InsightReport{
			Files:       m.Files,
			Description: m.Description,
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

// InsightAppReports godoc
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
// @description - insight.id
// @description - insight.name
// @description - insight.ruleset
// @description - insight.rule
// @description - insight.category
// @description - insight.effort
// @description - insight.labels
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
// @tags insightappreports
// @produce json
// @success 200 {object} []api.InsightAppReport
// @router /analyses/report/applications [get]
func (h AnalysisHandler) InsightAppReports(ctx *gin.Context) {
	resources := []InsightAppReport{}
	type M struct {
		ID                 uint
		Name               string
		Description        string
		BusinessService    string
		Effort             int
		Incidents          int
		Files              int
		InsightID          uint
		InsightName        string
		InsightDescription string
		RuleSet            string
		Rule               string
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
			{Field: "insight.id", Kind: qf.LITERAL},
			{Field: "insight.name", Kind: qf.LITERAL},
			{Field: "insight.ruleset", Kind: qf.STRING},
			{Field: "insight.rule", Kind: qf.STRING},
			{Field: "insight.category", Kind: qf.STRING},
			{Field: "insight.effort", Kind: qf.LITERAL},
			{Field: "insight.labels", Kind: qf.STRING, And: true},
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
		"i.ID InsightID",
		"i.Name InsightName",
		"i.RuleSet",
		"i.Rule")
	q = q.Table("Insight i")
	q = q.Joins("LEFT JOIN Incident n ON n.InsightID = i.ID")
	q = q.Joins("LEFT JOIN Analysis a ON a.ID = i.AnalysisID")
	q = q.Joins("LEFT JOIN Application app ON app.ID = a.ApplicationID")
	q = q.Joins("LEFT OUTER JOIN BusinessService b ON b.ID = app.BusinessServiceID")
	q = q.Where("a.ID IN (?)", h.analysisIDs(ctx, filter))
	q = q.Where("i.ID IN (?)", h.insightIDs(ctx, filter.Resource("insight")))
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
		r := InsightAppReport{}
		r.ID = m.ID
		r.Name = m.Name
		r.Description = m.Description
		r.BusinessService = m.BusinessService
		r.Effort = m.Effort
		r.Incidents = m.Incidents
		r.Files = m.Files
		r.Insight.ID = m.InsightID
		r.Insight.Name = m.InsightName
		r.Insight.Description = m.InsightName
		r.Insight.RuleSet = m.RuleSet
		r.Insight.Rule = m.Rule
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
// @router /analyses/report/insights/{id}/files [get]
// @param id path int true "Insight ID"
func (h AnalysisHandler) FileReports(ctx *gin.Context) {
	resources := []FileReport{}
	type M struct {
		InsightId uint
		File      string
		Effort    int
		Incidents int
	}
	// Insight
	insightId := h.pk(ctx)
	insight := &model.Insight{}
	result := h.DB(ctx).First(insight, insightId)
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
		"InsightId",
		"File",
		"Effort*COUNT(Incident.id) Effort",
		"COUNT(Incident.id) Incidents")
	q = q.Joins(",Insight")
	q = q.Where("Insight.ID = InsightID")
	q = q.Where("Insight.ID", insightId)
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
		r.InsightID = m.InsightId
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

// insightIDs returns insight filtered insight IDs.
// Filter:
//
//	insight.*
func (h *AnalysisHandler) insightIDs(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	q = q.Model(&model.Insight{})
	q = q.Select("ID")
	q = f.Where(q, "-Labels")
	filter := f
	if f, found := filter.Field("labels"); found {
		if f.Value.Operator(qf.AND) {
			var qs []*gorm.DB
			for _, f = range f.Expand() {
				f = f.As("json_each.value")
				iq := h.DB(ctx)
				iq = iq.Table("Insight")
				iq = iq.Joins("m ,json_each(Labels)")
				iq = iq.Select("m.ID")
				iq = f.Where(iq)
				qs = append(qs, iq)
			}
			q = q.Where("ID IN (?)", model.Intersect(qs...))
		} else {
			f = f.As("json_each.value")
			iq := h.DB(ctx)
			iq = iq.Table("Insight")
			iq = iq.Joins("m ,json_each(Labels)")
			iq = iq.Select("m.ID")
			iq = f.Where(iq)
			q = q.Where("ID IN (?)", iq)
		}
	}
	return
}

// depIDs returns insight filtered insight IDs.
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
// - Set the 'summary' field with archived insights.
// - Delete insights.
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
// - Set the 'summary' field with archived insights.
// - Delete insights.
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
// - Set the 'summary' field with archived insights.
// - Delete insights.
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
		db = db.Table("Insight i,")
		db = db.Joins("Incident n")
		db = db.Where("n.InsightID = i.ID")
		db = db.Where("i.AnalysisID", m.ID)
		db = db.Group("i.ID")
		summary := []model.ArchivedInsight{}
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
		err = db.Delete(&model.Insight{}).Error
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
	Application  Ref               `json:"application"`
	Effort       int               `json:"effort"`
	Commit       string            `json:"commit,omitempty" yaml:",omitempty"`
	Archived     bool              `json:"archived,omitempty" yaml:",omitempty"`
	Insights     []Insight         `json:"insights,omitempty" yaml:",omitempty"`
	Dependencies []TechDependency  `json:"dependencies,omitempty" yaml:",omitempty"`
	Summary      []ArchivedInsight `json:"summary,omitempty" yaml:",omitempty" swaggertype:"object"`
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
	m.Insights = []model.Insight{}
	return
}

// Insight REST resource.
type Insight struct {
	Resource    `yaml:",inline"`
	Analysis    uint       `json:"analysis"`
	RuleSet     string     `json:"ruleset" binding:"required"`
	Rule        string     `json:"rule" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty" yaml:",omitempty"`
	Category    string     `json:"category,omitempty" yaml:",omitempty"`
	Effort      int        `json:"effort,omitempty" yaml:",omitempty"`
	Incidents   []Incident `json:"incidents,omitempty" yaml:",omitempty"`
	Links       []Link     `json:"links,omitempty" yaml:",omitempty"`
	Facts       Map        `json:"facts,omitempty" yaml:",omitempty"`
	Labels      []string   `json:"labels"`
}

// With updates the resource with the model.
func (r *Insight) With(m *model.Insight) {
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
func (r *Insight) Model() (m *model.Insight) {
	m = &model.Insight{}
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
	Insight  uint   `json:"insight"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	CodeSnip string `json:"codeSnip" yaml:"codeSnip"`
	Facts    Map    `json:"facts"`
}

// With updates the resource with the model.
func (r *Incident) With(m *model.Incident) {
	r.Resource.With(&m.Model)
	r.Insight = m.InsightID
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

// ArchivedInsight created when insights are archived.
type ArchivedInsight model.ArchivedInsight

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

// InsightReport REST resource.
type InsightReport struct {
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

// InsightAppReport REST resource.
type InsightAppReport struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	BusinessService string `json:"businessService"`
	Effort          int    `json:"effort"`
	Incidents       int    `json:"incidents"`
	Files           int    `json:"files"`
	Insight         struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		RuleSet     string `json:"ruleset"`
		Rule        string `json:"rule"`
	} `json:"insight"`
}

// FileReport REST resource.
type FileReport struct {
	InsightID uint   `json:"insightId" yaml:"insightId"`
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

// InsightWriter used to create a file containing insights.
type InsightWriter struct {
	Encoder
	ctx *gin.Context
}

// Create an insights file and returns the path.
func (r *InsightWriter) Create(id uint, filter qf.Filter) (path string, count int64, err error) {
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
	file, err := os.CreateTemp("", "insight-*"+ext)
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
func (r *InsightWriter) db() (db *gorm.DB) {
	rtx := RichContext(r.ctx)
	db = rtx.DB.Debug()
	return
}

// Write the analysis file.
func (r *InsightWriter) Write(id uint, filter qf.Filter, output io.Writer) (count int64, err error) {
	r.Encoder, err = NewEncoder(r.ctx, output)
	if err != nil {
		return
	}
	page := Page{}
	page.With(r.ctx)
	sort := Sort{}
	err = sort.With(r.ctx, &model.Insight{})
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
		var insights []model.Insight
		err = db.Find(&insights).Error
		if err != nil {
			return
		}
		if len(insights) == 0 {
			break
		}
		for i := range insights {
			insight := Insight{}
			insight.With(&insights[i])
			r.writeItem(b, i, insight)
			count++
		}
	}
	r.endList()
	err = r.Encoder.error()
	return
}

// AnalysisWriter used to create a file containing an analysis.
type AnalysisWriter struct {
	Encoder
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
	r.Encoder, err = NewEncoder(r.ctx, output)
	if err != nil {
		return
	}
	r.begin()
	rx := &Analysis{}
	rx.With(m)
	r.embed(rx)
	err = r.addInsights(m)
	if err != nil {
		return
	}
	err = r.addDeps(m)
	if err != nil {
		return
	}
	r.end()
	err = r.Encoder.error()
	return
}

// addInsights writes insights (effort = 0).
func (r *AnalysisWriter) addInsights(m *model.Analysis) (err error) {
	r.field("insights")
	r.beginList()
	batch := 10
	for b := 0; ; b += batch {
		db := r.db()
		db = db.Preload("Incidents")
		db = db.Limit(batch)
		db = db.Offset(b)
		db = db.Where("AnalysisID", m.ID)
		var insights []model.Insight
		err = db.Find(&insights).Error
		if err != nil {
			return
		}
		if len(insights) == 0 {
			break
		}
		for i := range insights {
			insight := Insight{}
			insight.With(&insights[i])
			r.writeItem(b, i, insight)
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
	Encoder
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
	r.Encoder = &jsonEncoder{output: file}
	r.write("window[\"apps\"]=[")
	r.begin()
	r.field("id").writeStr(strconv.Itoa(int(m.Application.ID)))
	r.field("name").writeStr(m.Application.Name)
	r.field("analysis").writeStr(strconv.Itoa(int(m.ID)))
	err = r.addIssues(m)
	if err != nil {
		return
	}
	err = r.addInsights(m)
	if err != nil {
		return
	}
	err = r.addDeps(m)
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

// addIssues writes issues (effort > 0).
func (r *ReportWriter) addIssues(m *model.Analysis) (err error) {
	r.field("issues")
	r.beginList()
	batch := 10
	for b := 0; ; b += batch {
		db := r.db()
		db = db.Preload("Incidents")
		db = db.Limit(batch)
		db = db.Offset(b)
		db = db.Where("AnalysisID", m.ID)
		db = db.Where("effort > 0")
		var insights []model.Insight
		err = db.Find(&insights).Error
		if err != nil {
			return
		}
		if len(insights) == 0 {
			break
		}
		for i := range insights {
			insight := Insight{}
			insight.With(&insights[i])
			r.writeItem(b, i, insight)
		}
	}
	r.endList()
	return
}

// addInsights writes insights (effort = 0).
func (r *ReportWriter) addInsights(m *model.Analysis) (err error) {
	r.field("insights")
	r.beginList()
	batch := 10
	for b := 0; ; b += batch {
		db := r.db()
		db = db.Preload("Incidents")
		db = db.Limit(batch)
		db = db.Offset(b)
		db = db.Where("AnalysisID", m.ID)
		db = db.Where("effort == 0")
		var insights []model.Insight
		err = db.Find(&insights).Error
		if err != nil {
			return
		}
		if len(insights) == 0 {
			break
		}
		for i := range insights {
			insight := Insight{}
			insight.With(&insights[i])
			r.writeItem(b, i, insight)
		}
	}
	r.endList()
	return
}

// addDeps writes dependencies.
func (r *ReportWriter) addDeps(m *model.Analysis) (err error) {
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

// ManifestReader analysis manifest reader.
// The manifest contains 3 sections containing documents delimited by markers.
// The manifest must contain ALL markers even when sections are empty.
// Note: `^]` = `\x1D` = GS (group separator).
// Section markers:
//
//	^]BEGIN-MAIN^]
//	^]END-MAIN^]
//	^]BEGIN-INSIGHTS^]
//	^]END-INSIGHTS^]
//	^]BEGIN-DEPS^]
//	^]END-DEPS^]
type ManifestReader struct {
	sectionReader *io.SectionReader
	marker        map[string]Marker
	file          *os.File
}

// Open the reader delimited by the specified markers.
func (r *ManifestReader) Open(path, begin, end string) (err error) {
	err = r.scan(path)
	if err != nil {
		return
	}
	mBegin, found := r.marker[begin]
	if !found {
		err = &BadRequestError{
			Reason: fmt.Sprintf("marker: %s not found.", begin),
		}
		return
	}
	mEnd, found := r.marker[end]
	if !found {
		err = &BadRequestError{
			Reason: fmt.Sprintf("marker: %s not found.", end),
		}
		return
	}
	if mEnd.begin < mBegin.begin {
		err = &BadRequestError{
			Reason: fmt.Sprintf("marker: %s must precede %s.", begin, end),
		}
		return
	}
	r.file, err = os.Open(path)
	if err != nil {
		return
	}
	offset := mBegin.end
	n := mEnd.begin - offset
	r.sectionReader = io.NewSectionReader(r.file, offset, n)
	return
}

// Read bytes.
func (r *ManifestReader) Read(b []byte) (n int, err error) {
	if r.sectionReader == nil {
		err = io.EOF
		return
	}
	n, err = r.sectionReader.Read(b)
	return
}

// Close the reader.
func (r *ManifestReader) Close() (err error) {
	if r.file != nil {
		err = r.file.Close()
		r.file = nil
	}
	return
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
	offset := int64(0)
	var content []byte
	r.marker = make(map[string]Marker)
	reader := bufio.NewReaderSize(r.file, 1<<20)
	for {
		content, err = reader.ReadBytes('\n')
		if len(content) > 0 {
			token := strings.TrimSpace(string(content))
			token = strings.TrimRight(token, "\r\n")
			if pattern.MatchString(token) {
				m := Marker{}
				m.with(offset, content)
				r.marker[token] = m
			}
			offset += int64(len(content))
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			break
		}
	}
	return
}

// Marker manifest marker.
type Marker struct {
	begin int64
	end   int64
}

// with populates the marker.
func (m *Marker) with(offset int64, content []byte) {
	m.begin = offset
	m.end = offset
	m.end += int64(len(content))
}
