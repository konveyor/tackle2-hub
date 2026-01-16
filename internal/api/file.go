package api

import (
	"io"
	"mime"
	"net/http"
	"os"
	pathlib "path"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// FileHandler handles file routes.
type FileHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h FileHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("files"))
	routeGroup.GET(api.FilesRoute, h.List)
	routeGroup.GET(api.FilesRoute+"/", h.List)
	routeGroup.POST(api.FileRoute, h.Create)
	routeGroup.PUT(api.FileRoute, h.Create)
	routeGroup.PATCH(api.FileRoute, h.Append)
	routeGroup.GET(api.FileRoute, h.Get)
	routeGroup.DELETE(api.FileRoute, h.Delete)
}

// List godoc
// @summary List all files.
// @description List all files.
// @tags file
// @produce json
// @success 200 {object} []api.File
// @router /files [get]
func (h FileHandler) List(ctx *gin.Context) {
	var list []model.File
	result := h.DB(ctx).Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []File{}
	for i := range list {
		r := File{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a file.
// @description Create a file.
// @tags file
// @accept json
// @produce json
// @success 201 {object} api.File
// @router /files [post]
// @param name path string true "File name"
func (h FileHandler) Create(ctx *gin.Context) {
	m, err := h.create(ctx, ctx.Param(ID))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := File{}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// Append godoc
// @summary Append a file.
// @description Append a file.
// @tags file
// @accept json
// @produce json
// @success 204
// @router /files/{id} [put]
// @param id path uint true "File ID"
func (h FileHandler) Append(ctx *gin.Context) {
	var err error
	input, err := ctx.FormFile(FileField)
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
		_ = ctx.Error(err)
		return
	}
	m := &model.File{}
	id := h.pk(ctx)
	db := h.DB(ctx)
	err = db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	reader, err := input.Open()
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	writer, err := os.OpenFile(m.Path, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = writer.Close()
	}()
	_, err = io.Copy(writer, reader)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx)
	db = db.Model(m)
	user := h.BaseHandler.CurrentUser(ctx)
	err = db.Update("UpdateUser", user).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Get godoc
// @summary Get a file by ID.
// @description Get a file by ID. Returns api.File when Accept=application/json else the file content.
// @tags file
// @produce octet-stream
// @success 200 {object} api.File
// @router /files/{id} [get]
// @param id path int true "File ID"
func (h FileHandler) Get(ctx *gin.Context) {
	m := &model.File{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if h.Accepted(ctx, BindMIMEs...) {
		r := File{}
		r.With(m)
		h.Respond(ctx, http.StatusOK, r)
	} else {
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
// @tags file
// @success 204
// @router /files/{id} [delete]
// @param id path int true "File ID"
func (h FileHandler) Delete(ctx *gin.Context) {
	m := &model.File{}
	id := h.pk(ctx)
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.delete(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// create a file.
func (h FileHandler) create(ctx *gin.Context, name string) (m *model.File, err error) {
	mode := ctx.ContentType()
	switch mode {
	case binding.MIMEMultipartPOSTForm:
		m, err = h.createMultipart(ctx, name)
	case binding.MIMEYAML:
		m, err = h.createBody(ctx, name, binding.MIMEYAML)
	default:
		m, err = h.createBody(ctx, name, binding.MIMEJSON)
	}
	return
}

// create a file with multipart form.
func (h FileHandler) createMultipart(ctx *gin.Context, name string) (m *model.File, err error) {
	input, err := ctx.FormFile(FileField)
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
		return
	}
	m = &model.File{}
	m.Name = name
	m.Encoding = input.Header.Get(ContentType)
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Create(&m).Error
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			h.Status(ctx, http.StatusInternalServerError)
			_ = db.Delete(&m)
			return
		}
	}()
	reader, err := input.Open()
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
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
	return
}

// create a file with request body.
func (h FileHandler) createBody(ctx *gin.Context, name, encoding string) (m *model.File, err error) {
	m = &model.File{}
	m.Name = name
	m.Encoding = encoding
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Create(&m).Error
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			h.Status(ctx, http.StatusInternalServerError)
			_ = db.Delete(&m)
			return
		}
	}()
	reader := ctx.Request.Body
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
	return
}

// delete the specified file.
func (h FileHandler) delete(ctx *gin.Context, m *model.File) (err error) {
	err = os.Remove(m.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	}
	db := h.DB(ctx)
	err = db.Delete(m).Error
	return
}

// File REST resource.
type File = resource.File
