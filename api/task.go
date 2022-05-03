package api

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/tasking"
	"gorm.io/gorm/clause"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"path"
	"time"
)

//
// Routes
const (
	TasksRoot      = "/tasks"
	TaskRoot       = TasksRoot + "/:" + ID
	TaskReportRoot = TaskRoot + "/report"
	TaskBucketRoot = TaskRoot + "/bucket/*" + Wildcard
	TaskSubmitRoot = TaskRoot + "/submit"
)

const (
	LocatorParam = "locator"
)

//
// TaskHandler handles task routes.
type TaskHandler struct {
	BaseHandler
	BucketHandler
}

//
// AddRoutes adds routes.
func (h TaskHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, "tasks"))
	routeGroup.GET(TasksRoot, h.List)
	routeGroup.GET(TasksRoot+"/", h.List)
	routeGroup.POST(TasksRoot, h.Create)
	routeGroup.GET(TaskRoot, h.Get)
	routeGroup.PUT(TaskRoot, h.Update)
	routeGroup.DELETE(TaskRoot, h.Delete)
	routeGroup.PUT(TaskSubmitRoot, h.Submit, h.Update)
	routeGroup.GET(TaskBucketRoot, h.Content)
	routeGroup.POST(TaskBucketRoot, h.Upload)
	routeGroup.PUT(TaskBucketRoot, h.Upload)
	routeGroup.POST(TaskReportRoot, h.CreateReport)
	routeGroup.PUT(TaskReportRoot, h.UpdateReport)
	routeGroup.DELETE(TaskReportRoot, h.DeleteReport)
}

// Get godoc
// @summary Get a task by ID.
// @description Get a task by ID.
// @tags get
// @produce json
// @success 200 {object} api.Task
// @router /tasks/{id} [get]
// @param id path string true "Task ID"
func (h TaskHandler) Get(ctx *gin.Context) {
	task := &model.Task{}
	id := h.pk(ctx)
	db := h.DB.Preload(clause.Associations)
	result := db.First(task, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Task{}
	r.With(task)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all tasks.
// @description List all tasks.
// @tags get
// @produce json
// @success 200 {object} []api.Task
// @router /tasks [get]
func (h TaskHandler) List(ctx *gin.Context) {
	var list []model.Task
	db := h.DB
	locator := ctx.Query(LocatorParam)
	if locator != "" {
		db = db.Where("locator", locator)
	}
	db = db.Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Task{}
	for i := range list {
		r := Task{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a task.
// @description Create a task.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Task
// @router /tasks [post]
// @param task body api.Task true "Task data"
func (h TaskHandler) Create(ctx *gin.Context) {
	r := Task{}
	err := ctx.BindJSON(&r)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	switch r.State {
	case "":
		r.State = tasking.Created
	case tasking.Created,
		tasking.Ready:
	default:
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "state must be ('''|Created|Ready)",
			})
		return
	}
	m := r.Model()
	result := h.DB.Create(&m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a task.
// @description Delete a task.
// @tags delete
// @success 204
// @router /tasks/{id} [delete]
// @param id path string true "Task ID"
func (h TaskHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	task := &model.Task{}
	result := h.DB.First(task, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}
	if task.Pod != "" {
		pod := &core.Pod{}
		pod.Namespace = path.Dir(task.Pod)
		pod.Name = path.Base(task.Pod)
		err := h.Client.Delete(
			context.TODO(),
			pod)
		if err != nil {
			if !k8serr.IsNotFound(err) {
				h.deleteFailed(ctx, err)
				return
			}
		}
	}
	result = h.DB.Select(clause.Associations).Delete(task)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a task.
// @description Update a task.
// @tags update
// @accept json
// @success 204
// @router /tasks/{id} [put]
// @param id path string true "Task ID"
// @param task body Task true "Task data"
func (h TaskHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Task{}
	err := ctx.BindJSON(r)
	if err != nil {
		return
	}
	switch r.State {
	case tasking.Created,
		tasking.Ready:
	default:
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "state must be (Created|Ready)",
			})
		return
	}
	m := r.Model()
	m.Reset()
	db := h.DB.Model(m)
	db = db.Where("id", id)
	db = db.Where("state", tasking.Created)
	db = db.Omit("Bucket")
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Submit godoc
// @summary Submit a task.
// @description Submit a task.
// @tags update
// @accept json
// @success 202
// @router /tasks/{id}/submit [put]
// @param id path string true "Task ID"
func (h TaskHandler) Submit(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Task{}
	mod := func(withBody bool) (err error) {
		if !withBody {
			m := r.Model()
			err = h.DB.First(m, id).Error
			if err != nil {
				return
			}
			r.With(m)
		}
		r.State = tasking.Ready
		return
	}
	err := h.modBody(ctx, r, mod)
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}
	ctx.Next()
}

// Content godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @tags get
// @produce octet-stream
// @success 200
// @router /tasks/{id}/bucket/{wildcard} [get]
// @param id path string true "Task ID"
func (h TaskHandler) Content(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Task{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	h.content(ctx, &m.BucketOwner)
}

// Upload godoc
// @summary Upload bucket content by task ID and path.
// @description Upload bucket content by task ID and path.
// @tags get
// @produce json
// @success 204
// @router /tasks/{id}/bucket/{wildcard} [post]
// @param id path string true "Bucket ID"
func (h TaskHandler) Upload(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	h.upload(ctx, &m.BucketOwner)
}

// CreateReport godoc
// @summary Create a task report.
// @description Update a task report.
// @tags update
// @accept json
// @produce json
// @success 201 {object} api.TaskReport
// @router /tasks/{id}/report [post]
// @param id path string true "Task ID"
// @param task body api.TaskReport true "TaskReport data"
func (h TaskHandler) CreateReport(ctx *gin.Context) {
	id := h.pk(ctx)
	report := &TaskReport{}
	err := ctx.BindJSON(report)
	if err != nil {
		return
	}
	report.TaskID = id
	m := report.Model()
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
	}
	report.With(m)

	ctx.JSON(http.StatusCreated, report)
}

// UpdateReport godoc
// @summary Update a task report.
// @description Update a task report.
// @tags update
// @accept json
// @produce json
// @success 200 {object} api.TaskReport
// @router /tasks/{id}/report [put]
// @param id path string true "Task ID"
// @param task body api.TaskReport true "TaskReport data"
func (h TaskHandler) UpdateReport(ctx *gin.Context) {
	id := h.pk(ctx)
	report := &TaskReport{}
	err := ctx.BindJSON(report)
	if err != nil {
		return
	}
	report.TaskID = id
	m := report.Model()
	db := h.DB.Model(m)
	db = db.Where("taskid", id)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
	}
	report.With(m)

	ctx.JSON(http.StatusOK, report)
}

// DeleteReport godoc
// @summary Delete a task report.
// @description Delete a task report.
// @tags update
// @accept json
// @produce json
// @success 204
// @router /tasks/{id}/report [delete]
// @param id path string true "Task ID"
func (h TaskHandler) DeleteReport(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TaskReport{}
	m.ID = id
	db := h.DB.Where("taskid", id)
	result := db.Delete(&model.TaskReport{})
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Task REST resource.
type Task struct {
	Resource
	Name        string      `json:"name"`
	Locator     string      `json:"locator,omitempty"`
	Priority    int         `json:"priority,omitempty"`
	Variant     string      `json:"variant,omitempty"`
	Policy      string      `json:"policy,omitempty"`
	Addon       string      `json:"addon,omitempty" binding:"required"`
	Data        interface{} `json:"data" swaggertype:"object" binding:"required"`
	Application *Ref        `json:"application,omitempty"`
	State       string      `json:"state"`
	Image       string      `json:"image,omitempty"`
	Bucket      string      `json:"bucket,omitempty"`
	Purged      bool        `json:"purged,omitempty"`
	Started     *time.Time  `json:"started,omitempty"`
	Terminated  *time.Time  `json:"terminated,omitempty"`
	Error       string      `json:"error,omitempty"`
	Pod         string      `json:"pod,omitempty"`
	Retries     int         `json:"retries,omitempty"`
	Report      *TaskReport `json:"report,omitempty"`
}

//
// With updates the resource with the model.
func (r *Task) With(m *model.Task) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Image = m.Image
	r.Addon = m.Addon
	r.Locator = m.Locator
	r.Priority = m.Priority
	r.Policy = m.Policy
	r.Variant = m.Variant
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Bucket = m.Bucket
	r.State = m.State
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Error = m.Error
	r.Pod = m.Pod
	r.Retries = m.Retries
	_ = json.Unmarshal(m.Data, &r.Data)
	if m.Report != nil {
		report := &TaskReport{}
		report.With(m.Report)
		r.Report = report
	}
}

//
// Model builds a model.
func (r *Task) Model() (m *model.Task) {
	m = &model.Task{
		Name:          r.Name,
		Addon:         r.Addon,
		Locator:       r.Locator,
		Variant:       r.Variant,
		Priority:      r.Priority,
		Policy:        r.Policy,
		State:         r.State,
		Retries:       r.Retries,
		ApplicationID: r.idPtr(r.Application),
	}
	m.Bucket = r.Bucket
	m.Data, _ = json.Marshal(r.Data)
	m.ID = r.ID
	return
}

//
// TaskReport REST resource.
type TaskReport struct {
	Resource
	Status    string      `json:"status"`
	Error     string      `json:"error"`
	Total     int         `json:"total"`
	Completed int         `json:"completed"`
	Activity  []string    `json:"activity"`
	Result    interface{} `json:"result,omitempty"`
	TaskID    uint        `json:"task"`
}

//
// With updates the resource with the model.
func (r *TaskReport) With(m *model.TaskReport) {
	r.Resource.With(&m.Model)
	r.Status = m.Status
	r.Error = m.Error
	r.Total = m.Total
	r.Completed = m.Completed
	r.TaskID = m.TaskID
	if m.Activity != nil {
		_ = json.Unmarshal(m.Activity, &r.Activity)
	}
	if m.Result != nil {
		_ = json.Unmarshal(m.Result, &r.Result)
	}
}

//
// Model builds a model.
func (r *TaskReport) Model() (m *model.TaskReport) {
	if r.Activity == nil {
		r.Activity = []string{}
	}
	m = &model.TaskReport{
		Status:    r.Status,
		Error:     r.Error,
		Total:     r.Total,
		Completed: r.Completed,
		TaskID:    r.TaskID,
	}
	if r.Activity != nil {
		m.Activity, _ = json.Marshal(r.Activity)
	}
	if r.Result != nil {
		m.Result, _ = json.Marshal(r.Result)
	}
	m.ID = r.ID

	return
}
