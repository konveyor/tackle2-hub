package api

import (
	"bytes"
	"encoding/csv"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Record types
const (
	RecordTypeApplication = "1"
	RecordTypeDependency  = "2"
)

const (
	ExpectedFieldCount = 17
)

// ImportHandler handles import routes.
type ImportHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h ImportHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("imports"))
	routeGroup.GET(api.SummariesRoute, h.ListSummaries)
	routeGroup.GET(api.SummariesRoute+"/", h.ListSummaries)
	routeGroup.GET(api.SummaryRoute, h.GetSummary)
	routeGroup.DELETE(api.SummaryRoute, h.DeleteSummary)
	routeGroup.GET(api.ImportsRoute, h.ListImports)
	routeGroup.GET(api.ImportsRoute+"/", h.ListImports)
	routeGroup.GET(api.ImportRoute, h.GetImport)
	routeGroup.DELETE(api.ImportRoute, h.DeleteImport)
	routeGroup.GET(api.DownloadRoute, h.DownloadCSV)
	routeGroup.POST(api.UploadRoute, h.UploadCSV)
}

// GetImport godoc
// @summary Get an import by ID.
// @description Get an import by ID.
// @tags imports
// @produce json
// @success 200 {object} api.Import
// @router /imports/{id} [get]
// @param id path int true "Import ID"
func (h ImportHandler) GetImport(ctx *gin.Context) {
	m := &model.Import{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB(ctx), "ImportTags")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	h.Respond(ctx, http.StatusOK, m.AsMap())
}

// ListImports godoc
// @summary List imports.
// @description List imports.
// @tags imports
// @produce json
// @success 200 {object} []api.Import
// @router /imports [get]
func (h ImportHandler) ListImports(ctx *gin.Context) {
	var list []model.Import
	db := h.DB(ctx)
	summaryId := ctx.Query("importSummary.id")
	if summaryId != "" {
		db = db.Where("importsummaryid = ?", summaryId)
	}
	isValid := ctx.Query("isValid")
	if isValid == "true" {
		db = db.Where("isvalid")
	} else if isValid == "false" {
		db = db.Not("isvalid")
	}
	db = h.preLoad(db, "ImportTags")
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Import{}
	for i := range list {
		resources = append(resources, list[i].AsMap())
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// DeleteImport godoc
// @summary Delete an import.
// @description Delete an import. This leaves any created application or dependency.
// @tags imports
// @success 204
// @router /imports/{id} [delete]
// @param id path int true "Import ID"
func (h ImportHandler) DeleteImport(ctx *gin.Context) {
	id := ctx.Param(ID)
	result := h.DB(ctx).Delete(&model.Import{}, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// GetSummary godoc
// @summary Get an import summary by ID.
// @description Get an import by ID.
// @tags imports
// @produce json
// @success 200 {object} api.ImportSummary
// @router /importsummaries/{id} [get]
// @param id path int true "ImportSummary ID"
func (h ImportHandler) GetSummary(ctx *gin.Context) {
	m := &model.ImportSummary{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB(ctx), "Imports")
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	h.Respond(ctx, http.StatusOK, m)
}

// ListSummaries godoc
// @summary List import summaries.
// @description List import summaries.
// @tags imports
// @produce json
// @success 200 {object} []api.ImportSummary
// @router /importsummaries [get]
func (h ImportHandler) ListSummaries(ctx *gin.Context) {
	var list []model.ImportSummary
	db := h.preLoad(h.DB(ctx), "Imports")
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []ImportSummary{}
	for i := range list {
		r := ImportSummary{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// DeleteSummary godoc
// @summary Delete an import summary and associated import records.
// @description Delete an import summary and associated import records.
// @tags imports
// @success 204
// @router /importsummaries/{id} [delete]
// @param id path int true "ImportSummary ID"
func (h ImportHandler) DeleteSummary(ctx *gin.Context) {
	id := ctx.Param(ID)
	result := h.DB(ctx).Delete(&model.ImportSummary{}, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// UploadCSV godoc
// @summary Upload a CSV containing applications and dependencies to import.
// @description Upload a CSV containing applications and dependencies to import.
// @tags imports
// @success 201 {object} api.ImportSummary
// @produce json
// @router /importsummaries/upload [post]
func (h ImportHandler) UploadCSV(ctx *gin.Context) {
	fileName, ok := ctx.GetPostForm("fileName")
	if !ok {
		h.Status(ctx, http.StatusBadRequest)
	}
	file, err := ctx.FormFile(FileField)
	if err != nil {
		h.Status(ctx, http.StatusBadRequest)
	}
	fileReader, err := file.Open()
	if err != nil {
		h.Status(ctx, http.StatusBadRequest)
	}
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, fileReader)
	if err != nil {
		h.Status(ctx, http.StatusInternalServerError)
	}
	createEntitiesField, ok := ctx.GetPostForm("createEntities")
	if !ok {
		createEntitiesField = "1"
	}
	createEntities, err := strconv.ParseBool(createEntitiesField)
	if err != nil {
		createEntities = true
	}
	m := model.ImportSummary{
		Filename:       fileName,
		ImportStatus:   resource.InProgress,
		Content:        buf.Bytes(),
		CreateEntities: createEntities,
	}
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(&m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	_, err = fileReader.Seek(0, 0)
	if err != nil {
		h.Status(ctx, http.StatusInternalServerError)
	}
	csvReader := csv.NewReader(fileReader)
	csvReader.TrimLeadingSpace = true
	// skip the header
	_, err = csvReader.Read()
	if err != nil {
		h.Status(ctx, http.StatusBadRequest)
	}

	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				h.Status(ctx, http.StatusBadRequest)
			}
		}
		var imp model.Import
		switch row[0] {
		case RecordTypeApplication:
			// Check row format - length, expecting 17 fields + tags
			if len(row) < ExpectedFieldCount {
				h.Respond(ctx, http.StatusBadRequest, gin.H{"errorMessage": "Invalid Application Import CSV format."})
				return
			}
			imp = h.applicationFromRow(fileName, row)
		case RecordTypeDependency:
			imp = h.dependencyFromRow(fileName, row)
		default:
			imp = model.Import{
				Filename:    fileName,
				RecordType1: row[0],
			}
		}
		imp.ImportSummary = m
		result := h.DB(ctx).Create(&imp)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
	}

	summary := ImportSummary{}
	summary.With(&m)
	h.Respond(ctx, http.StatusCreated, summary)
}

// DownloadCSV godoc
// @summary Export the source CSV for a particular import summary.
// @description Export the source CSV for a particular import summary.
// @tags imports
// @produce text/csv
// @success 200 file csv
// @router /importsummaries/download [get]
// @param importSummary.id query string true "ImportSummary ID"
func (h ImportHandler) DownloadCSV(ctx *gin.Context) {
	id := ctx.Query("importSummary.id")
	m := &model.ImportSummary{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	h.Attachment(ctx, m.Filename)
	ctx.Data(http.StatusOK, "text/csv", m.Content)
}

// CSV upload supports two types of records in the same file: application imports, and dependencies.
// A dependency row must consist of the following columns:
//
// Col 1: Record Type 1 -- This will always contain a "2" for a dependency
// Col 2: Application Name -- The name of the application that has the dependency relationship.
//
//	This application must exist.
//
// Col 6: Dependency -- The name of the application on the other side of the dependency relationship.
// Col 7: Dependency Direction -- Whether this is a "northbound" or "southbound" dependency.
//
// Between the Application Name and the Dependency field there may be an arbitrary number of columns representing
// tags or other fields that only pertain to an application import. The dependency and direction will always be
// the last two columns in the row.
//
// Examples:
//
// 2,MyApplication,,,,OtherApplication,NORTHBOUND,,,,,,,,
func (h ImportHandler) dependencyFromRow(fileName string, row []string) (app model.Import) {
	app = model.Import{
		Filename:            fileName,
		RecordType1:         row[0],
		ApplicationName:     row[1],
		Dependency:          row[5],
		DependencyDirection: row[6],
	}
	return
}

// CSV upload supports two types of records in the same file: application imports, and dependencies.
// An application row must consist of the following columns:
//
// Col 1: Record Type 1 -- This will always contain a "1" for an application
// Col 2: Application Name -- The name of the application to be created.
// Col 3: Description -- A short description of the application.
// Col 4: Comments -- Additional comments on the application.
// Col 5: Business Service -- The name of the business service this Application should belong to.
//
//	This business service must already exist.
//
// Col 6: Dependency -- Optional dependency to another Application (by name)
// Col 7: Dependency direction -- Either northbound or southbound
//
// Binary: Binary coordinates (like from <Group>:<Artifact>:<Version>:<Packaging>).
// Col 8: Group
// Col 9: Artifact
// Col 10: Version
// Col 11: Packaging (optional)
//
// Repository: The following columns are coordinates to a source repository.
// Col 12: Kind (defaults to 'git' if empty)
// Col 13: URL
// Col 14: Branch
// Col 15: Path
//
// Following that are up to twenty pairs of Tag Categories and Tags, specified by name. These are optional.
// If a tag category and a tag are specified, they must already exist.
//
// Examples:
//
// 1,MyApplication,My cool app,No comment,Marketing,,,binarygrp,elfbin,v1,war,git,url,branch,path,TagType1,Tag1,TagType2,Tag2
// 1,OtherApplication,,,Marketing,MyApplication,southbound
func (h ImportHandler) applicationFromRow(fileName string, row []string) (app model.Import) {
	app = model.Import{
		Filename:            fileName,
		RecordType1:         row[0],
		ApplicationName:     row[1],
		Description:         row[2],
		Comments:            row[3],
		BusinessService:     row[4],
		Dependency:          row[5],
		DependencyDirection: row[6],
		BinaryGroup:         row[7],
		BinaryArtifact:      row[8],
		BinaryVersion:       row[9],
		BinaryPackaging:     row[10],
		RepositoryKind:      row[11],
		RepositoryURL:       row[12],
		RepositoryBranch:    row[13],
		RepositoryPath:      row[14],
		Owner:               row[15],
		Contributors:        row[16],
	}

	// Tags
	for i := ExpectedFieldCount; i < len(row); i++ {
		if i%2 == 0 {
			tag := model.ImportTag{
				Name:     row[i],
				Category: row[i-1],
			}
			app.ImportTags = append(app.ImportTags, tag)
		}
	}

	return
}

// Import REST resource.
type Import map[string]any

// ImportSummary REST resource.
type ImportSummary = resource.ImportSummary
