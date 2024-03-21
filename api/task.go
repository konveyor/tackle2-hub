package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/strings/slices"
)

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

// TaskHandler handles task routes.
type TaskHandler struct {
	BucketOwner
}

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
// @param id path int true "Task ID"
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
	q := ctx.Query("merged")
	if b, _ := strconv.ParseBool(q); b {
		err := r.injectFiles(h.DB(ctx))
		if err != nil {
			_ = ctx.Error(result.Error)
			return
		}
	}

	h.Respond(ctx, http.StatusOK, r)
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
	db := h.DB(ctx)
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

	h.Respond(ctx, http.StatusOK, resources)
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
		h.Respond(ctx,
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

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a task.
// @description Delete a task.
// @tags tasks
// @success 204
// @router /tasks/{id} [delete]
// @param id path int true "Task ID"
func (h TaskHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	task := &model.Task{}
	result := h.DB(ctx).First(task, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	rt := tasking.Task{Task: task}
	err := rt.Delete(h.Client(ctx))
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

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a task.
// @description Update a task.
// @tags tasks
// @accept json
// @success 204
// @router /tasks/{id} [put]
// @param id path int true "Task ID"
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
		h.Respond(ctx,
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

	h.Status(ctx, http.StatusNoContent)
}

// Submit godoc
// @summary Submit a task.
// @description Submit a task.
// @tags tasks
// @accept json
// @success 204
// @router /tasks/{id}/submit [put]
// @param id path int true "Task ID"
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
// @param id path int true "Task ID"
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
		h.Respond(ctx,
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

	h.Status(ctx, http.StatusNoContent)
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
// @param id path int true "Task ID"
// @param wildcard path string true "Content path"
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
		h.Status(ctx, http.StatusNotFound)
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
// @param id path int true "Task ID"
// @param wildcard path string true "Content path"
func (h TaskHandler) BucketPut(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		h.Status(ctx, http.StatusNotFound)
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
// @param id path int true "Task ID"
// @param wildcard path string true "Content path"
func (h TaskHandler) BucketDelete(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if !m.HasBucket() {
		h.Status(ctx, http.StatusNotFound)
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
// @param id path int true "Task ID"
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

	h.Respond(ctx, http.StatusCreated, report)
}

// UpdateReport godoc
// @summary Update a task report.
// @description Update a task report.
// @tags tasks
// @accept json
// @produce json
// @success 200 {object} api.TaskReport
// @router /tasks/{id}/report [put]
// @param id path int true "Task ID"
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

	h.Respond(ctx, http.StatusOK, report)
}

// DeleteReport godoc
// @summary Delete a task report.
// @description Delete a task report.
// @tags tasks
// @accept json
// @produce json
// @success 204
// @router /tasks/{id}/report [delete]
// @param id path int true "Task ID"
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

	h.Status(ctx, http.StatusNoContent)
}

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

// TTL time-to-live.
type TTL struct {
	Created   int `json:"created,omitempty"`
	Pending   int `json:"pending,omitempty"`
	Postponed int `json:"postponed,omitempty"`
	Running   int `json:"running,omitempty"`
	Succeeded int `json:"succeeded,omitempty"`
	Failed    int `json:"failed,omitempty"`
}

// TaskError used in Task.Errors.
type TaskError struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// Task REST resource.
type Task struct {
	Resource    `yaml:",inline"`
	Name        string       `json:"name"`
	Locator     string       `json:"locator,omitempty" yaml:",omitempty"`
	Priority    int          `json:"priority,omitempty" yaml:",omitempty"`
	Policy      string       `json:"policy,omitempty" yaml:",omitempty"`
	TTL         *TTL         `json:"ttl,omitempty" yaml:",omitempty"`
	Kind        string       `json:"kind,omitempty" yaml:",omitempty"`
	Addon       string       `json:"addon,omitempty" yaml:",omitempty"`
	Extensions  []string     `json:"extensions,omitempty" yaml:",omitempty"`
	Data        interface{}  `json:"data" swaggertype:"object" binding:"required"`
	Application *Ref         `json:"application,omitempty" yaml:",omitempty"`
	State       string       `json:"state"`
	Image       string       `json:"image,omitempty" yaml:",omitempty"`
	Pod         string       `json:"pod,omitempty" yaml:",omitempty"`
	Retries     int          `json:"retries,omitempty" yaml:",omitempty"`
	Started     *time.Time   `json:"started,omitempty" yaml:",omitempty"`
	Terminated  *time.Time   `json:"terminated,omitempty" yaml:",omitempty"`
	Canceled    bool         `json:"canceled,omitempty" yaml:",omitempty"`
	Bucket      *Ref         `json:"bucket,omitempty" yaml:",omitempty"`
	Purged      bool         `json:"purged,omitempty" yaml:",omitempty"`
	Errors      []TaskError  `json:"errors,omitempty" yaml:",omitempty"`
	Activity    []string     `json:"activity,omitempty" yaml:",omitempty"`
	Attached    []Attachment `json:"attached" yaml:",omitempty"`
}

// With updates the resource with the model.
func (r *Task) With(m *model.Task) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.Locator = m.Locator
	r.Priority = m.Priority
	r.Policy = m.Policy
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.State = m.State
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Pod = m.Pod
	r.Retries = m.Retries
	r.Canceled = m.Canceled
	_ = json.Unmarshal(m.Data, &r.Data)
	if m.TTL != nil {
		_ = json.Unmarshal(m.TTL, &r.TTL)
	}
	if m.Extensions != nil {
		_ = json.Unmarshal(m.Extensions, &r.Extensions)
	}
	if m.Errors != nil {
		_ = json.Unmarshal(m.Errors, &r.Errors)
	}
	if m.Attached != nil {
		_ = json.Unmarshal(m.Attached, &r.Attached)
	}
	if m.Report != nil {
		report := &TaskReport{}
		report.With(m.Report)
		r.Activity = report.Activity
		r.Errors = append(r.Errors, report.Errors...)
		r.Attached = append(r.Attached, report.Attached...)
		switch r.State {
		case tasking.Succeeded:
			switch report.Status {
			case tasking.Failed:
				r.State = report.Status
			}
		}
	}
}

// Model builds a model.
func (r *Task) Model() (m *model.Task) {
	m = &model.Task{
		Name:          r.Name,
		Addon:         r.Addon,
		Kind:          r.Kind,
		Locator:       r.Locator,
		Priority:      r.Priority,
		Policy:        r.Policy,
		State:         r.State,
		ApplicationID: r.idPtr(r.Application),
	}
	m.Data, _ = json.Marshal(StrMap(r.Data))
	m.ID = r.ID
	if r.TTL != nil {
		m.TTL, _ = json.Marshal(r.TTL)
	}
	if r.Extensions != nil {
		m.Extensions, _ = json.Marshal(r.Extensions)
	}
	return
}

// injectFiles inject attached files into the activity.
func (r *Task) injectFiles(db *gorm.DB) (err error) {
	sort.Slice(
		r.Attached,
		func(i, j int) bool {
			return r.Attached[i].Activity > r.Attached[j].Activity
		})
	for _, ref := range r.Attached {
		if ref.Activity == 0 {
			continue
		}
		if ref.Activity > len(r.Activity) {
			continue
		}
		m := &model.File{}
		err = db.First(m, ref.ID).Error
		if err != nil {
			return
		}
		b, nErr := ioutil.ReadFile(m.Path)
		if nErr != nil {
			err = nErr
			return
		}
		var content []string
		for _, s := range strings.Split(string(b), "\n") {
			content = append(
				content,
				"> "+s)
		}
		snipA := slices.Clone(r.Activity[:ref.Activity])
		snipB := slices.Clone(r.Activity[ref.Activity:])
		r.Activity = append(
			append(snipA, content...),
			snipB...)
	}
	return
}

// TaskReport REST resource.
type TaskReport struct {
	Resource  `yaml:",inline"`
	Status    string       `json:"status"`
	Errors    []TaskError  `json:"errors,omitempty" yaml:",omitempty"`
	Total     int          `json:"total,omitempty" yaml:",omitempty"`
	Completed int          `json:"completed,omitempty" yaml:",omitempty"`
	Activity  []string     `json:"activity,omitempty" yaml:",omitempty"`
	Attached  []Attachment `json:"attached,omitempty" yaml:",omitempty"`
	Result    interface{}  `json:"result,omitempty" yaml:",omitempty" swaggertype:"object"`
	TaskID    uint         `json:"task"`
}

// With updates the resource with the model.
func (r *TaskReport) With(m *model.TaskReport) {
	r.Resource.With(&m.Model)
	r.Status = m.Status
	r.Total = m.Total
	r.Completed = m.Completed
	r.TaskID = m.TaskID
	if m.Activity != nil {
		_ = json.Unmarshal(m.Activity, &r.Activity)
	}
	if m.Errors != nil {
		_ = json.Unmarshal(m.Errors, &r.Errors)
	}
	if m.Attached != nil {
		_ = json.Unmarshal(m.Attached, &r.Attached)
	}
	if m.Result != nil {
		_ = json.Unmarshal(m.Result, &r.Result)
	}
}

// Model builds a model.
func (r *TaskReport) Model() (m *model.TaskReport) {
	if r.Activity == nil {
		r.Activity = []string{}
	}
	m = &model.TaskReport{
		Status:    r.Status,
		Total:     r.Total,
		Completed: r.Completed,
		TaskID:    r.TaskID,
	}
	if r.Activity != nil {
		m.Activity, _ = json.Marshal(r.Activity)
	}
	if r.Result != nil {
		m.Result, _ = json.Marshal(StrMap(r.Result))
	}
	if r.Errors != nil {
		m.Errors, _ = json.Marshal(r.Errors)
	}
	if r.Attached != nil {
		m.Attached, _ = json.Marshal(r.Attached)
	}
	m.ID = r.ID

	return
}

// Attachment associates Files with a TaskReport.
type Attachment struct {
	// Ref references an attached File.
	Ref `yaml:",inline"`
	// Activity index (1-based) association with an
	// activity entry. Zero(0) indicates not associated.
	Activity int `json:"activity,omitempty" yaml:",omitempty"`
}
