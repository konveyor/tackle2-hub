package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"io"
	"net/http"
	"os"
	pathlib "path"
)

//
// Routes
const (
	BucketsRoot          = "/buckets"
	BucketRoot           = BucketsRoot + "/:" + ID
	BucketContent        = BucketRoot + "/content/*" + Wildcard
	AppBucketsRoot       = ApplicationRoot + BucketsRoot
	AppBucketRoot        = AppBucketsRoot + "/:" + Name
	AppBucketContentRoot = AppBucketRoot + "/content/*" + Wildcard
)

//
// BucketHandler handles bucket routes.
type BucketHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h BucketHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, "buckets"))
	routeGroup.GET(BucketsRoot, h.List)
	routeGroup.GET(BucketsRoot+"/", h.List)
	routeGroup.POST(BucketsRoot, h.Create)
	routeGroup.GET(BucketRoot, h.Get)
	routeGroup.DELETE(BucketRoot, h.Delete)
	routeGroup.GET(BucketContent, h.GetContent)
	routeGroup.POST(BucketContent, h.UploadContent)
	routeGroup.PUT(BucketContent, h.UploadContent)
	routeGroup.GET(AppBucketsRoot, h.AppList)
	routeGroup.GET(AppBucketsRoot+"/", h.AppList)
	routeGroup.GET(AppBucketRoot+"/", h.AppGet)
	routeGroup.POST(AppBucketRoot, h.AppCreate)
	routeGroup.GET(AppBucketContentRoot, h.AppContent)
	routeGroup.POST(AppBucketContentRoot, h.AppUploadContent)
	routeGroup.PUT(AppBucketContentRoot, h.AppUploadContent)
}

// Get godoc
// @summary Get a bucket by ID.
// @description Get a bucket by ID.
// @tags get
// @produce json
// @success 200 {object} Bucket
// @router /buckets/{id} [get]
// @param id path string true "Bucket ID"
func (h BucketHandler) Get(ctx *gin.Context) {
	m := &model.Bucket{}
	id := ctx.Param(ID)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Bucket{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all buckets.
// @description List all buckets.
// @tags get
// @produce json
// @success 200 {object} []Bucket
// @router /buckets [get]
func (h BucketHandler) List(ctx *gin.Context) {
	var list []model.Bucket
	result := h.DB.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Bucket{}
	for i := range list {
		r := Bucket{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a bucket.
// @description Create a bucket.
// @tags create
// @accept json
// @produce json
// @success 201 {object} Bucket
// @router /buckets [post]
// @param bucket body Bucket true "Bucket data"
func (h BucketHandler) Create(ctx *gin.Context) {
	r := &Bucket{}
	err := ctx.BindJSON(r)
	if err != nil {
		return
	}
	err = h.create(r)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a bucket.
// @description Delete a bucket.
// @tags delete
// @success 204
// @router /buckets/{id} [delete]
// @param id path string true "Bucket ID"
func (h BucketHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	m := &model.Bucket{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}
	result = h.DB.Delete(m, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetContent godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @tags get
// @produce octet-stream
// @success 200
// @router /bucket/{id}/content/{wildcard} [get]
// @param id path string true "Bucket ID"
func (h BucketHandler) GetContent(ctx *gin.Context) {
	rPath := ctx.Param(Wildcard)
	m := &model.Bucket{}
	id := ctx.Param(ID)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	ctx.File(pathlib.Join(
		m.Path,
		rPath))
}

// UploadContent godoc
// @summary Upload bucket content by ID and path.
// @description Upload bucket content by ID and path.
// @tags get
// @produce json
// @success 204
// @router /bucket/{id}/content/{wildcard} [post]
// @param id path string true "Bucket ID"
func (h BucketHandler) UploadContent(ctx *gin.Context) {
	m := &model.Bucket{}
	id := ctx.Param(ID)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := &Bucket{}
	r.With(m)
	h.upload(ctx, r)
}

// AppList godoc
// @summary List buckets associated with application.
// @description List buckets associated with application.
// @tags get
// @produce json
// @success 200 {object} []Bucket
// @router /applications/{id}/buckets [get]
// @param id path int true "Application ID"
func (h BucketHandler) AppList(ctx *gin.Context) {
	var list []model.Bucket
	appId := ctx.Param(ID)
	db := h.DB.Where("applicationid", appId)
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Bucket{}
	for i := range list {
		r := Bucket{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// AppGet godoc
// @summary List buckets associated with application.
// @description List buckets associated with application.
// @tags get
// @produce json
// @success 200 {object} Bucket
// @router /applications/{id}/buckets/{name} [get]
// @param id path int true "Application ID"
// @param name path string true "Bucket Name"
func (h BucketHandler) AppGet(ctx *gin.Context) {
	m := &model.Bucket{}
	appId := ctx.Param(ID)
	name := ctx.Param(Name)
	db := h.DB.Where("applicationID", appId).Where("name", name)
	result := db.First(m)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Bucket{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// AppCreate godoc
// @summary Create a bucket for an application.
// @description Create a bucket for an application.
// @tags create
// @accept json
// @produce json
// @success 201 {object} Bucket
// @router /applications/{id}/buckets/{name} [post]
// @param id path int true "Application ID"
// @param name path string true "Bucket Name"
// @param bucket body Bucket true "Bucket data"
func (h BucketHandler) AppCreate(ctx *gin.Context) {
	appID := ctx.Param(ID)
	name := ctx.Param(Name)
	application := &model.Application{}
	result := h.DB.First(application, appID)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	r := &Bucket{}
	r.Application.ID = application.ID
	r.Name = name
	err := h.create(r)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, r)
}

// AppContent godoc
// @summary Get bucket content by application ID, bucket name and path.
// @description Get bucket content by application ID, bucket name and path.
// @tags get
// @produce octet-stream
// @success 200
// @router /applications/{id}/buckets/{name}/content/{wildcard} [get]
// @param id path string true "Bucket ID"
// @param name path string true "Bucket Name"
func (h BucketHandler) AppContent(ctx *gin.Context) {
	rPath := ctx.Param(Wildcard)
	appID := ctx.Param(ID)
	name := ctx.Param(Name)
	m := &model.Bucket{}
	db := h.DB.Where("applicationID", appID).Where("name", name)
	result := db.First(m)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	ctx.File(pathlib.Join(
		m.Path,
		rPath))
}

// AppUploadContent godoc
// @summary Upload bucket content by application ID, bucket name and path.
// @description Upload bucket content by application ID, bucket name and path.
// @tags get
// @produce json
// @success 204
// @router /applications/{id}/buckets/{name}/content/{wildcard} [post]
// @param id path string true "Bucket ID"
// @param name path string true "Bucket Name"
func (h BucketHandler) AppUploadContent(ctx *gin.Context) {
	appID := ctx.Param(ID)
	name := ctx.Param(Name)
	m := &model.Bucket{}
	db := h.DB.Where("applicationID", appID).Where("name", name)
	result := db.First(m)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := &Bucket{}
	r.With(m)
	h.upload(ctx, r)
}

//
// create a bucket.
func (h *BucketHandler) create(r *Bucket) (err error) {
	uid := uuid.New()
	r.Path = pathlib.Join(
		Settings.Hub.Bucket.Path,
		uid.String())
	err = os.MkdirAll(r.Path, 0777)
	if err != nil {
		return
	}

	m := r.Model()
	result := h.DB.Create(m)
	err = result.Error
	if err != nil {
		_ = os.Remove(r.Path)
	}
	r.With(m)

	return
}

//
// upload file.
func (h *BucketHandler) upload(ctx *gin.Context, b *Bucket) {
	rPath := ctx.Param(Wildcard)
	path := pathlib.Join(
		b.Path,
		rPath)
	input, err := ctx.FormFile("file")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	defer func() {
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
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
	if err != nil {
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Bucket REST Resource.
type Bucket struct {
	Resource
	Name        string `json:"name" binding:"alphanum|containsany=_-"`
	Path        string `json:"path"`
	Application Ref    `json:"application" binding:"required"`
}

//
// With updates the resource with the model.
func (r *Bucket) With(m *model.Bucket) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Path = m.Path
	r.Application.ID = m.ApplicationID
}

//
// Model builds a model.
func (r *Bucket) Model() (m *model.Bucket) {
	m = &model.Bucket{
		Name:          r.Name,
		Path:          r.Path,
		ApplicationID: r.Application.ID,
	}
	m.ID = r.ID

	return
}
