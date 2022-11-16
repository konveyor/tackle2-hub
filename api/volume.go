package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	VolumesRoot     = "/volumes"
	VolumeRoot      = VolumesRoot + "/:" + ID
	VolumeCleanRoot = VolumeRoot + "/clean"
)

//
// VolumeHandler handles volume routes.
type VolumeHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h VolumeHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("volumes"))
	routeGroup.GET(VolumesRoot, h.List)
	routeGroup.GET(VolumesRoot+"/", h.List)
	routeGroup.GET(VolumeRoot, h.Get)
	routeGroup.PUT(VolumeRoot, h.Update)
	routeGroup.POST(VolumeCleanRoot, h.Clean)
}

// Get godoc
// @summary Get an volume by ID.
// @description Get an volume by ID.
// @tags get
// @produce json
// @success 200 {object} api.Volume
// @router /volumes/{id} [get]
// @param name path string true "Volume ID"
func (h VolumeHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Volume{}
	err := h.DB.First(m, id).Error
	if err != nil {
		h.getFailed(ctx, err)
		return
	}
	r := Volume{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all volumes.
// @description List all volumes.
// @tags get
// @produce json
// @success 200 {object} []api.Volume
// @router /volumes [get]
func (h VolumeHandler) List(ctx *gin.Context) {
	list := []model.Volume{}
	err := h.DB.Find(&list).Error
	if err != nil {
		h.listFailed(ctx, err)
		return
	}
	resources := []Volume{}
	for i := range list {
		r := Volume{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Update godoc
// @summary Update a volume.
// @description Update a volume.
// @tags update
// @accept json
// @success 204
// @router /volumes/{id} [put]
// @param id path string true "Volume ID"
// @param job_function body api.Volume true "Volume data"
func (h VolumeHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Volume{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB.Model(m)
	db = db.Omit(clause.Associations)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Clean godoc
// @summary Clean the volume.
// @description Clean the volume.
// @tags post
// @produce json
// @success 202 {object} api.Task
// @router /volumes [post]
func (h VolumeHandler) Clean(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Volume{}
	err := h.DB.First(m, id).Error
	if err != nil {
		h.getFailed(ctx, err)
		return
	}
	r := Task{}
	r.Variant = "mount:clean"
	r.Name = r.Variant
	r.Locator = r.Variant
	r.Addon = "admin"
	r.State = task.Ready
	r.Priority = 1
	r.TTL = &TTL{
		Running:   10,
		Succeeded: 1,
		Failed:    10,
	}
	r.Policy = task.Isolated
	r.Data = map[string][]interface{}{
		"volumes": {m.ID},
	}
	t := r.Model()
	err = h.DB.Create(t).Error
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	r.With(t)
	ctx.JSON(http.StatusAccepted, r)
}

//
// Volume REST resource.
type Volume struct {
	Resource
	Name     string `json:"name"`
	Capacity string `json:"capacity"`
	Used     string `json:"used"`
}

//
// With model.
func (r *Volume) With(m *model.Volume) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Capacity = m.Capacity
	r.Used = m.Used
}

//
// Model builds a model.
func (r *Volume) Model() (m *model.Volume) {
	m = &model.Volume{
		Name:     r.Name,
		Capacity: r.Capacity,
		Used:     r.Used,
	}
	m.ID = r.ID

	return
}
