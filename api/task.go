package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"time"
)

//
// Routes
const (
	TasksRoot             = "/tasks"
	TaskRoot              = TasksRoot + "/:" + ID
	TaskReportRoot        = TaskRoot + "/report"
	TaskBucketRoot        = TaskRoot + "/bucket"
	TaskBucketContentRoot = TaskBucketRoot + "/*" + Wildcard
	TaskSubmitRoot        = TaskRoot + "/submit"
	TaskCancelRoot        = TaskRoot + "/cancel"
)

const (
	LocatorParam = "locator"
)

//
// TaskHandler handles task routes.
type TaskHandler struct {
	BucketOwner
}

//
// AddRoutes adds routes.
func (h TaskHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tasks"))
	routeGroup.GET(TasksRoot, h.List)
	routeGroup.GET(TasksRoot+"/", h.List)
	routeGroup.POST(TasksRoot, h.Create)
	routeGroup.GET(TaskRoot, h.Get)
	routeGroup.PUT(TaskRoot, h.Update)
	routeGroup.DELETE(TaskRoot, h.Delete)
	// Actions
	routeGroup.PUT(TaskSubmitRoot, h.Submit, h.Update)
	routeGroup.PUT(TaskCancelRoot, h.Cancel)
	// Bucket
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tasks.bucket"))
	routeGroup.GET(TaskBucketRoot, h.BucketGet)
	routeGroup.GET(TaskBucketContentRoot, h.BucketGet)
	routeGroup.POST(TaskBucketContentRoot, h.BucketPut)
	routeGroup.PUT(TaskBucketContentRoot, h.BucketPut)
	routeGroup.DELETE(TaskBucketContentRoot, h.BucketDelete)
	// Report
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tasks.report"))
	routeGroup.POST(TaskReportRoot, h.CreateReport)
	routeGroup.PUT(TaskReportRoot, h.UpdateReport)
	routeGroup.DELETE(TaskReportRoot, h.DeleteReport)
}

// Get godoc
// @summary Get a task by ID.
// @description Get a task by ID.
// @tags tasks
// @produce json
// @success 200 {object} api.Task
// @router /tasks/{id} [get]
// @param id path string true "Task ID"
func (h TaskHandler) Get(ctx *gin.Context) {
	task := &model.Task{}
	id := h.pk(ctx)
	db := h.DB(ctx).Preload(clause.Associations)
	result := db.First(task, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Task{}
	r.With(task)

	h.Render(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all tasks.
// @description List all tasks.
// @tags tasks
// @produce json
// @success 200 {object} []api.Task
// @router /tasks [get]
func (h TaskHandler) List(ctx *gin.Context) {
	var list []model.Task
	db := h.Paginated(ctx)
	locator := ctx.Query(LocatorParam)
	if locator != "" {
		db = db.Where("locator", locator)
	}
	db = db.Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Task{}
	for i := range list {
		r := Task{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a task.
// @description Create a task.
// @tags tasks
// @accept json
// @produce json
// @success 201 {object} api.Task
// @router /tasks [post]
// @param task body api.Task true "Task data"
func (h TaskHandler) Create(ctx *gin.Context) {
	r := Task{}
	err := h.Bind(ctx, &r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	switch r.State {
	case "":
		r.State = tasking.Created
	case tasking.Created,
		tasking.Ready:
	default:
		h.Render(ctx,
			http.StatusBadRequest,
			gin.H{
				"error": "state must be (''|Created|Ready)",
			})
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(&m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

	h.Render(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a task.
// @description Delete a task.
// @tags tasks
// @success 204
// @router /tasks/{id} [delete]
// @param id path string true "Task ID"
func (h TaskHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	task := &model.Task{}
	result := h.DB(ctx).First(task, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	rt := tasking.Task{Task: task}
	err := rt.Delete(h.Client)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			_ = ctx.Error(err)
			return
		}
	}
	result = h.DB(ctx).Delete(task)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update a task.
// @description Update a task.
// @tags tasks
// @accept json
// @success 204
// @router /tasks/{id} [put]
// @param id path string true "Task ID"
// @param task body Task true "Task data"
func (h TaskHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Task{}
	err := h.Bind(ctx, r)
	if err != nil {
		return
	}
	switch r.State {
	case tasking.Created,
		tasking.Ready:
	default:
		h.Render(ctx,
			http.StatusBadRequest,
			gin.H{
				"error": "state must be (Created|Ready)",
			})
		return
	}
	m := r.Model()
	m.Reset()
	db := h.DB(ctx).Model(m)
	db = db.Where("id", id)
	db = db.Where("state", tasking.Created)
	db = h.omitted(db)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Submit godoc
// @summary Submit a task.
// @description Submit a task.
// @tags tasks
// @accept json
// @success 204
// @router /tasks/{id}/submit [put]
// @param id path string true "Task ID"
// @param task body Task false "Task data (optional)"
func (h TaskHandler) Submit(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Task{}
	mod := func(withBody bool) (err error) {
		if !withBody {
			m := r.Model()
			err = h.DB(ctx).First(m, id).Error
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
		_ = ctx.Error(err)
		return
	}
	ctx.Next()
}

// Cancel godoc
// @summary Cancel a task.
// @description Cancel a task.
// @tags tasks
// @success 204
// @router /tasks/{id}/cancel [put]
// @param id path string true "Task ID"
func (h TaskHandler) Cancel(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Task{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	switch m.State {
	case tasking.Succeeded,
		tasking.Failed,
		tasking.Canceled:
		h.Render(ctx,
			http.StatusBadRequest,
			gin.H{
				"error": "state must not be (Succeeded|Failed|Canceled)",
			})
		return
	}
	db := h.DB(ctx).Model(m)
	db = db.Where("id", id)
	db = db.Where(
		"state not IN ?",
		[]string{
			tasking.Succeeded,
			tasking.Failed,
			tasking.Canceled,
		})
	err := db.Update("Canceled", true).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// BucketGet godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @description Returns index.html for directories when Accept=text/html else a tarball.
// @description ?filter=glob supports directory content filtering.
// @tags tasks
// @produce octet-stream
// @success 200
// @router /tasks/{id}/bucket/{wildcard} [get]
// @param id path string true "Task ID"
// @param filter query string false "Filter"
func (h TaskHandler) BucketGet(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		ctx.Status(http.StatusNotFound)
		return
	}

	h.bucketGet(ctx, *m.BucketID)
}

// BucketPut godoc
// @summary Upload bucket content by ID and path.
// @description Upload bucket content by ID and path (handles both [post] and [put] requests).
// @tags tasks
// @produce json
// @success 204
// @router /tasks/{id}/bucket/{wildcard} [post]
// @param id path string true "Task ID"
func (h TaskHandler) BucketPut(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		ctx.Status(http.StatusNotFound)
		return
	}

	h.bucketPut(ctx, *m.BucketID)
}

// BucketDelete godoc
// @summary Delete bucket content by ID and path.
// @description Delete bucket content by ID and path.
// @tags tasks
// @produce json
// @success 204
// @router /tasks/{id}/bucket/{wildcard} [delete]
// @param id path string true "Task ID"
func (h TaskHandler) BucketDelete(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		ctx.Status(http.StatusNotFound)
		return
	}

	h.bucketDelete(ctx, *m.BucketID)
}

// CreateReport godoc
// @summary Create a task report.
// @description Update a task report.
// @tags tasks
// @accept json
// @produce json
// @success 201 {object} api.TaskReport
// @router /tasks/{id}/report [post]
// @param id path string true "Task ID"
// @param task body api.TaskReport true "TaskReport data"
func (h TaskHandler) CreateReport(ctx *gin.Context) {
	id := h.pk(ctx)
	report := &TaskReport{}
	err := h.Bind(ctx, report)
	if err != nil {
		return
	}
	report.TaskID = id
	m := report.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}
	report.With(m)

	h.Render(ctx, http.StatusCreated, report)
}

// UpdateReport godoc
// @summary Update a task report.
// @description Update a task report.
// @tags tasks
// @accept json
// @produce json
// @success 200 {object} api.TaskReport
// @router /tasks/{id}/report [put]
// @param id path string true "Task ID"
// @param task body api.TaskReport true "TaskReport data"
func (h TaskHandler) UpdateReport(ctx *gin.Context) {
	id := h.pk(ctx)
	report := &TaskReport{}
	err := h.Bind(ctx, report)
	if err != nil {
		return
	}
	report.TaskID = id
	m := report.Model()
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Where("taskid", id)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}
	report.With(m)

	h.Render(ctx, http.StatusOK, report)
}

// DeleteReport godoc
// @summary Delete a task report.
// @description Delete a task report.
// @tags tasks
// @accept json
// @produce json
// @success 204
// @router /tasks/{id}/report [delete]
// @param id path string true "Task ID"
func (h TaskHandler) DeleteReport(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TaskReport{}
	m.ID = id
	db := h.DB(ctx).Where("taskid", id)
	result := db.Delete(&model.TaskReport{})
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Fields omitted by:
//   - Create
//   - Update.
func (h *TaskHandler) omitted(db *gorm.DB) (out *gorm.DB) {
	out = db.Omit([]string{
		"BucketID",
		"Bucket",
		"Image",
		"Pod",
		"Started",
		"Terminated",
		"Canceled",
		"Error",
		"Retries",
	}...)
	return
}

//
// TTL time-to-live.
type TTL struct {
	Created   int `json:"created,omitempty"`
	Pending   int `json:"pending,omitempty"`
	Postponed int `json:"postponed,omitempty"`
	Running   int `json:"running,omitempty"`
	Succeeded int `json:"succeeded,omitempty"`
	Failed    int `json:"failed,omitempty"`
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
	TTL         *TTL        `json:"ttl,omitempty"`
	Addon       string      `json:"addon,omitempty" binding:"required"`
	Data        interface{} `json:"data" swaggertype:"object" binding:"required"`
	Application *Ref        `json:"application,omitempty"`
	State       string      `json:"state"`
	Image       string      `json:"image,omitempty"`
	Bucket      *Ref        `json:"bucket,omitempty"`
	Purged      bool        `json:"purged,omitempty"`
	Started     *time.Time  `json:"started,omitempty"`
	Terminated  *time.Time  `json:"terminated,omitempty"`
	Error       string      `json:"error,omitempty"`
	Pod         string      `json:"pod,omitempty"`
	Retries     int         `json:"retries,omitempty"`
	Canceled    bool        `json:"canceled,omitempty"`
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
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.State = m.State
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Error = m.Error
	r.Pod = m.Pod
	r.Retries = m.Retries
	r.Canceled = m.Canceled
	_ = json.Unmarshal(m.Data, &r.Data)
	if m.Report != nil {
		report := &TaskReport{}
		report.With(m.Report)
		r.Report = report
	}
	if m.TTL != nil {
		_ = json.Unmarshal(m.TTL, &r.TTL)
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
		ApplicationID: r.idPtr(r.Application),
	}
	m.Data, _ = json.Marshal(r.Data)
	m.ID = r.ID
	if r.TTL != nil {
		m.TTL, _ = json.Marshal(r.TTL)
	}
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
	Result    interface{} `json:"result,omitempty" swaggertype:"object"`
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
