package api

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/nas"
	"io"
	"net/http"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
	"time"
)

//
// Routes
const (
	BucketsRoot       = "/buckets"
	BucketRoot        = BucketsRoot + "/:" + ID
	BucketContentRoot = BucketRoot + "/*" + Wildcard
)

//
// Params
const (
	Filter = "filter"
)

//
// BucketHandler handles bucket routes.
type BucketHandler struct {
	BucketOwner
}

//
// AddRoutes adds routes.
func (h BucketHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("buckets"))
	routeGroup.GET(BucketsRoot, h.List)
	routeGroup.GET(BucketsRoot+"/", h.List)
	routeGroup.POST(BucketsRoot, h.Create)
	routeGroup.GET(BucketRoot, h.Get)
	routeGroup.DELETE(BucketRoot, h.Delete)
	routeGroup.POST(BucketContentRoot, h.BucketPut)
	routeGroup.PUT(BucketContentRoot, h.BucketPut)
	routeGroup.GET(BucketContentRoot, h.BucketGet)
	routeGroup.DELETE(BucketContentRoot, h.BucketDelete)
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
	result := h.Paginated(ctx).Find(&list)
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

	h.Render(ctx, http.StatusOK, resources)
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
	h.Render(ctx, http.StatusCreated, r)
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
// @param id path string true "Bucket ID"
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
		h.Render(ctx, http.StatusOK, r)
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
// @param id path string true "Bucket ID"
func (h BucketHandler) Delete(ctx *gin.Context) {
	m := &model.Bucket{}
	id := h.pk(ctx)
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = os.Remove(m.Path)
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

	ctx.Status(http.StatusNoContent)
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
// @param id path string true "Task ID"
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
// @param id path string true "Bucket ID"
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
// @param id path string true "Bucket ID"
func (h BucketHandler) BucketDelete(ctx *gin.Context) {
	h.bucketDelete(ctx, h.pk(ctx))
}

//
// Bucket REST resource.
type Bucket struct {
	Resource
	Path       string     `json:"path"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

//
// With updates the resource with the model.
func (r *Bucket) With(m *model.Bucket) {
	r.Resource.With(&m.Model)
	r.Path = m.Path
	r.Expiration = m.Expiration
}

type BucketOwner struct {
	BaseHandler
}

//
// bucketGet reads bucket content.
// When path is DIRECTORY:
//    Accept=text/html return body is index.html.
//    Else streams tarball.
// When path is FILE:
//    Streams FILE content.
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
		ctx.Status(http.StatusNotFound)
		return
	}
	if st.IsDir() {
		filter := DirFilter{
			pattern: ctx.Query(Filter),
			root:    path,
		}
		if h.Accepted(ctx, binding.MIMEHTML) {
			err = h.getFile(ctx, m)
			if err != nil {
				_ = ctx.Error(err)
			}
			return
		} else {
			err := h.getDir(ctx, path, filter)
			if err != nil {
				_ = ctx.Error(err)
				return
			}
			ctx.Status(http.StatusOK)
		}
	} else {
		err = h.getFile(ctx, m)
		if err != nil {
			_ = ctx.Error(err)
		}
		return
	}
}

//
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
	ctx.Status(http.StatusNoContent)
}

//
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
	ctx.Status(http.StatusNoContent)
}

//
// putDir write a directory into bucket.
func (h *BucketOwner) putDir(ctx *gin.Context, output string) (err error) {
	file, err := ctx.FormFile(FileField)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	fileReader, err := file.Open()
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer func() {
		_ = fileReader.Close()
	}()
	zipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer func() {
		_ = zipReader.Close()
	}()
	err = nas.RmDir(output)
	if err != nil {
		return
	}
	err = os.MkdirAll(output, 0777)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	tarReader := tar.NewReader(zipReader)
	for {
		header, nErr := tarReader.Next()
		if nErr != nil {
			if nErr == io.EOF {
				break
			} else {
				err = liberr.Wrap(nErr)
				return
			}
		}
		switch header.Typeflag {
		case tar.TypeDir:
			path := pathlib.Join(output, header.Name)
			err = os.Mkdir(path, 0777)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		case tar.TypeReg:
			path := pathlib.Join(output, header.Name)
			file, nErr := os.Create(path)
			if nErr != nil {
				err = liberr.Wrap(nErr)
				return
			}
			_, err = io.Copy(file, tarReader)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			_ = file.Close()
		}
	}
	return
}

//
// getDir reads a directory from the bucket.
func (h *BucketOwner) getDir(ctx *gin.Context, input string, filter DirFilter) (err error) {
	var tarOutput bytes.Buffer
	tarWriter := tar.NewWriter(&tarOutput)
	err = filepath.Walk(
		input,
		func(path string, info os.FileInfo, wErr error) (err error) {
			if wErr != nil {
				err = liberr.Wrap(wErr)
				return
			}
			if path == input {
				return
			}
			if !filter.Match(path) {
				return
			}
			header, err := tar.FileInfoHeader(info, path)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			header.Name = strings.Replace(path, input, "", 1)
			switch header.Typeflag {
			case tar.TypeDir:
				err = tarWriter.WriteHeader(header)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
			case tar.TypeReg:
				err = tarWriter.WriteHeader(header)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
				file, nErr := os.Open(path)
				if err != nil {
					err = liberr.Wrap(nErr)
					return
				}
				defer func() {
					_ = file.Close()
				}()
				_, err = io.Copy(tarWriter, file)
				if err != nil {
					err = liberr.Wrap(err)
					return
				}
			}
			return
		})
	if err != nil {
		return
	}
	err = tarWriter.Close()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	ctx.Writer.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", pathlib.Base(input)+".tar.gz"))
	ctx.Writer.Header().Set(Directory, DirectoryExpand)
	zipReader := bufio.NewReader(&tarOutput)
	zipWriter := gzip.NewWriter(ctx.Writer)
	defer func() {
		_ = zipWriter.Close()
	}()
	_, err = io.Copy(zipWriter, zipReader)
	return
}

//
// getFile reads a file from the bucket.
func (h *BucketOwner) getFile(ctx *gin.Context, m *model.Bucket) (err error) {
	rPath := ctx.Param(Wildcard)
	path := pathlib.Join(m.Path, rPath)
	ctx.File(path)
	return
}

//
// putFile writes a file to the bucket.
func (h *BucketOwner) putFile(ctx *gin.Context, m *model.Bucket) (err error) {
	path := pathlib.Join(m.Path, ctx.Param(Wildcard))
	input, err := ctx.FormFile(FileField)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	reader, err := input.Open()
	if err != nil {
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

//
// DirFilter supports glob-style filtering.
type DirFilter struct {
	root    string
	pattern string
	cache   map[string]bool
}

//
// Match determines if path matches the filter.
func (r *DirFilter) Match(path string) (b bool) {
	if r.pattern == "" {
		b = true
		return
	}
	if r.cache == nil {
		r.cache = map[string]bool{}
		matches, _ := filepath.Glob(pathlib.Join(r.root, r.pattern))
		for _, p := range matches {
			r.cache[p] = true
		}
	}
	_, b = r.cache[path]
	return
}
