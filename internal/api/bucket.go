package api

import (
	"io"
	"net/http"
	"os"
	pathlib "path"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"github.com/konveyor/tackle2-hub/shared/tar"
)

// BucketHandler handles bucket routes.
type BucketHandler struct {
	BucketOwner
}

// AddRoutes adds routes.
func (h BucketHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("buckets"))
	routeGroup.GET(api.BucketsRoute, h.List)
	routeGroup.GET(api.BucketsRoute+"/", h.List)
	routeGroup.POST(api.BucketsRoute, h.Create)
	routeGroup.GET(api.BucketRoute, h.Get)
	routeGroup.DELETE(api.BucketRoute, h.Delete)
	routeGroup.POST(api.BucketContentRoute, h.BucketPut)
	routeGroup.PUT(api.BucketContentRoute, h.BucketPut)
	routeGroup.GET(api.BucketContentRoute, h.BucketGet)
	routeGroup.DELETE(api.BucketContentRoute, h.BucketDelete)
}

// List godoc
// @summary List all buckets.
// @description List all buckets.
// @tags buckets
// @produce json
// @success 200 {object} []api.Bucket
// @router /buckets [get]
func (h BucketHandler) List(ctx *gin.Context) {
	var list []model.Bucket
	result := h.DB(ctx).Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Bucket{}
	for i := range list {
		r := Bucket{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a bucket.
// @description Create a bucket.
// @tags buckets
// @accept json
// @produce json
// @success 201 {object} api.Bucket
// @router /buckets [post]
// @param name path string true "Bucket name"
func (h BucketHandler) Create(ctx *gin.Context) {
	m := &model.Bucket{}
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(&m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Bucket{}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// Get godoc
// @summary Get a bucket by ID.
// @description Get a bucket by ID.
// @description Returns api.Bucket when Accept=application/json.
// @description Else returns index.html when Accept=text/html.
// @description Else returns tarball.
// @tags buckets
// @produce octet-stream
// @success 200 {object} api.Bucket
// @router /buckets/{id} [get]
// @param id path int true "Bucket ID"
func (h BucketHandler) Get(ctx *gin.Context) {
	m := &model.Bucket{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if h.Accepted(ctx, BindMIMEs...) {
		r := Bucket{}
		r.With(m)
		h.Respond(ctx, http.StatusOK, r)
		return
	}
	h.bucketGet(ctx, id)
}

// Delete godoc
// @summary Delete a bucket.
// @description Delete a bucket.
// @tags buckets
// @success 204
// @router /buckets/{id} [delete]
// @param id path int true "Bucket ID"
func (h BucketHandler) Delete(ctx *gin.Context) {
	m := &model.Bucket{}
	id := h.pk(ctx)
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = nas.RmDir(m.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			_ = ctx.Error(err)
			return
		}
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// BucketGet godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @description When path is FILE, returns file content.
// @description When path is DIRECTORY and Accept=text/html returns index.html.
// @description ?filter=glob supports directory content filtering.
// @description Else returns a tarball.
// @tags buckets
// @produce octet-stream
// @success 200
// @router /buckets/{id}/{wildcard} [get]
// @param id path int true "Task ID"
// @param wildcard path string true "Content path"
// @param filter query string false "Filter"
func (h BucketHandler) BucketGet(ctx *gin.Context) {
	h.bucketGet(ctx, h.pk(ctx))
}

// BucketPut godoc
// @summary Upload bucket content by ID and path.
// @description Upload bucket content by ID and path (handles both [post] and [put] requests).
// @tags buckets
// @produce json
// @success 204
// @router /buckets/{id}/{wildcard} [post]
// @param id path int true "Bucket ID"
// @param wildcard path string true "Content path"
func (h BucketHandler) BucketPut(ctx *gin.Context) {
	h.bucketPut(ctx, h.pk(ctx))
}

// BucketDelete godoc
// @summary Delete bucket content by ID and path.
// @description Delete bucket content by ID and path.
// @tags buckets
// @produce json
// @success 204
// @router /buckets/{id}/{wildcard} [delete]
// @param id path int true "Bucket ID"
// @param wildcard path string true "Content path"
func (h BucketHandler) BucketDelete(ctx *gin.Context) {
	h.bucketDelete(ctx, h.pk(ctx))
}

// Bucket REST resource.
type Bucket = resource.Bucket

type BucketOwner struct {
	BaseHandler
}

// bucketGet reads bucket content.
// When path is DIRECTORY:
//
//	Accept=text/html return body is index.html.
//	Else streams tarball.
//
// When path is FILE:
//
//	Streams FILE content.
func (h *BucketOwner) bucketGet(ctx *gin.Context, id uint) {
	var err error
	m := &model.Bucket{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	path := pathlib.Join(m.Path, ctx.Param(Wildcard))
	st, err := os.Stat(path)
	if os.IsNotExist(err) {
		h.Status(ctx, http.StatusNotFound)
		return
	}
	if st.IsDir() {
		filter := tar.NewFilter(path)
		filter.Include(ctx.Query(Filter))
		if h.Accepted(ctx, binding.MIMEHTML) {
			h.getFile(ctx, m)
		} else {
			h.getDir(ctx, path, filter)
		}
	} else {
		h.getFile(ctx, m)
	}
}

// bucketPut write a file to the bucket.
// The `Directory` header determines how the uploaded file is to be handled.
// When `Directory`=Expand, the file (TARBALL) is extracted into the bucket.
// Else the file is stored.
func (h *BucketOwner) bucketPut(ctx *gin.Context, id uint) {
	var err error
	m := &model.Bucket{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if ctx.Request.Header.Get(Directory) == DirectoryExpand {
		err = h.putDir(ctx, pathlib.Join(m.Path, ctx.Param(Wildcard)))
	} else {
		err = h.putFile(ctx, m)
	}
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// bucketDelete content from the bucket.
func (h *BucketOwner) bucketDelete(ctx *gin.Context, id uint) {
	m := &model.Bucket{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	rPath := ctx.Param(Wildcard)
	path := pathlib.Join(m.Path, rPath)
	err := nas.RmDir(path)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Status(ctx, http.StatusNoContent)
}

// putDir write a directory into bucket.
func (h *BucketOwner) putDir(ctx *gin.Context, output string) (err error) {
	file, err := ctx.FormFile(FileField)
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
		return
	}
	fileReader, err := file.Open()
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = fileReader.Close()
	}()
	err = nas.RmDir(output)
	if err != nil {
		return
	}
	tarReader := tar.NewReader()
	err = tarReader.Extract(output, fileReader)
	return
}

// getDir reads a directory from the bucket.
func (h *BucketOwner) getDir(ctx *gin.Context, input string, filter tar.Filter) {
	tarWriter := tar.NewWriter(ctx.Writer)
	tarWriter.Filter = filter
	defer func() {
		tarWriter.Close()
	}()
	err := tarWriter.AssertDir(input)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Attachment(ctx, pathlib.Base(input)+".tar.gz")
	ctx.Writer.Header().Set(Directory, DirectoryExpand)
	ctx.Status(http.StatusOK)
	_ = tarWriter.AddDir(input)
	return
}

// getFile reads a file from the bucket.
func (h *BucketOwner) getFile(ctx *gin.Context, m *model.Bucket) {
	rPath := ctx.Param(Wildcard)
	path := pathlib.Join(m.Path, rPath)
	ctx.File(path)
}

// putFile writes a file to the bucket.
func (h *BucketOwner) putFile(ctx *gin.Context, m *model.Bucket) (err error) {
	path := pathlib.Join(m.Path, ctx.Param(Wildcard))
	input, err := ctx.FormFile(FileField)
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
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
	err = os.MkdirAll(pathlib.Dir(path), 0777)
	if err != nil {
		return
	}
	writer, err := os.Create(path)
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
	err = os.Chmod(path, 0666)
	return
}
