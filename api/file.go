package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"io"
	"mime"
	"net/http"
	"os"
	pathlib "path"
	"time"
)

//
// Routes
const (
	FilesRoot = "/files"
	FileRoot  = FilesRoot + "/:" + ID
)

//
// FileHandler handles file routes.
type FileHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h FileHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("files"))
	routeGroup.GET(FilesRoot, h.List)
	routeGroup.GET(FilesRoot+"/", h.List)
	routeGroup.POST(FileRoot, h.Create)
	routeGroup.PUT(FileRoot, h.Create)
	routeGroup.GET(FileRoot, h.Get)
	routeGroup.DELETE(FileRoot, h.Delete)
}

// List godoc
// @summary List all files.
// @description List all files.
// @tags get
// @produce json
// @success 200 {object} []api.File
// @router /files [get]
func (h FileHandler) List(ctx *gin.Context) {
	var list []model.File
	result := h.DB.Find(&list)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}
	resources := []File{}
	for i := range list {
		r := File{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a file.
// @description Create a file.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.File
// @router /files [post]
// @param name path string true "File name"
func (h FileHandler) Create(ctx *gin.Context) {
	var err error
	input, err := ctx.FormFile(FileField)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	m := &model.File{}
	m.Name = ctx.Param(ID)
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB.Create(&m)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}
	defer func() {
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			_ = h.DB.Delete(&m)
			return
		}
	}()
	reader, err := input.Open()
	if err != nil {
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	writer, err := os.Create(m.Path)
	if err != nil {
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	_, err = io.Copy(writer, reader)
	if err != nil {
		return
	}
	err = os.Chmod(m.Path, 0666)
	if err != nil {
		return
	}
	r := File{}
	r.With(m)
	ctx.JSON(http.StatusCreated, r)
}

// Get godoc
// @summary Get a file by ID.
// @description Get a file by ID. Returns api.File when Accept=application/json else the file content.
// @tags get
// @produce json octet-stream
// @success 200 {object} api.File
// @router /files/{id} [get]
// @param id path string true "File ID"
func (h FileHandler) Get(ctx *gin.Context) {
	m := &model.File{}
	id := h.pk(ctx)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.reportError(ctx, result.Error)
		return
	}
	switch ctx.GetHeader(Accept) {
	case AppJson:
		r := File{}
		r.With(m)
		ctx.JSON(http.StatusOK, r)
	default:
		header := ctx.Writer.Header()
		header[ContentType] = []string{
			mime.TypeByExtension(pathlib.Ext(m.Name)),
		}
		ctx.File(m.Path)
	}
}

// Delete godoc
// @summary Delete a file.
// @description Delete a file.
// @tags delete
// @success 204
// @router /files/{id} [delete]
// @param id path string true "File ID"
func (h FileHandler) Delete(ctx *gin.Context) {
	m := &model.File{}
	id := h.pk(ctx)
	err := h.DB.First(m, id).Error
	if err != nil {
		h.reportError(ctx, err)
		return
	}
	err = os.Remove(m.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			h.reportError(ctx, err)
			return
		}
	}
	err = h.DB.Delete(m).Error
	if err != nil {
		h.reportError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// File REST resource.
type File struct {
	Resource
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

//
// With updates the resource with the model.
func (r *File) With(m *model.File) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Path = m.Path
	r.Expiration = m.Expiration
}
