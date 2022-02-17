package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konveyor/tackle2-hub/model"
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
	e.GET(BucketsRoot, h.List)
	e.GET(BucketsRoot+"/", h.List)
	e.POST(BucketsRoot, h.Create)
	e.GET(BucketRoot, h.Get)
	e.DELETE(BucketRoot, h.Delete)
	e.GET(BucketContent, h.GetContent)
	e.GET(AppBucketsRoot, h.AppList)
	e.GET(AppBucketsRoot+"/", h.AppList)
	e.GET(AppBucketRoot+"/", h.AppGet)
	e.POST(AppBucketRoot, h.AppCreate)
	e.GET(AppBucketContentRoot, h.AppContent)
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
	pagination := NewPagination(ctx)
	db := pagination.apply(h.DB)
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
// @success 204 {object} Bucket
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
// @produce json
// @success 200 {object}
// @router /bucket/{id}/content/* [get]
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

// AppList godoc
// @summary List buckets associated with application.
// @description List buckets associated with application.
// @tags get
// @produce json
// @success 200 {object} []Bucket
// @router /application-inventory/application/{id}/buckets [get]
// @param id path int true "Application ID"
func (h BucketHandler) AppList(ctx *gin.Context) {
	var list []model.Bucket
	appId := ctx.Param(ID)
	pagination := NewPagination(ctx)
	db := pagination.apply(h.DB)
	db = db.Where("applicationid", appId)
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
// @router /application-inventory/application/{id}/buckets/{name} [get]
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
// @router /application-inventory/application/{id}/buckets/{name} [post]
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
	r.ApplicationID = application.ID
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
// @produce json
// @success 200 {object}
// @router /application-inventory/application/{id}/buckets/{name}/content/* [get]
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

//
// create a bucket.
func (h BucketHandler) create(r *Bucket) (err error) {
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
// Bucket REST Resource.
type Bucket struct {
	Resource
	Name          string `json:"name" binding:"alphanum|containsany=_-"`
	Path          string `json:"path"`
	ApplicationID uint   `json:"application" binding:"required"`
}

//
// With updates the resource with the model.
func (r *Bucket) With(m *model.Bucket) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Path = m.Path
	r.ApplicationID = m.ApplicationID
}

//
// Model builds a model.
func (r *Bucket) Model() (m *model.Bucket) {
	m = &model.Bucket{
		Name:          r.Name,
		Path:          r.Path,
		ApplicationID: r.ApplicationID,
	}
	m.ID = r.ID

	return
}
