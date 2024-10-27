package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/tar"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/utils/strings/slices"
)

// Routes
const (
	TasksRoot                = "/tasks"
	TasksReportRoot          = TasksRoot + "/report"
	TasksReportQueueRoot     = TasksReportRoot + "/queue"
	TasksReportQueueRootByIds=TasksRoot+"/multiple"
	TasksReportDashboardRoot = TasksReportRoot + "/dashboard"
	TaskRoot                 = TasksRoot + "/:" + ID
	TaskReportRoot           = TaskRoot + "/report"
	TaskAttachedRoot         = TaskRoot + "/attached"
	TaskBucketRoot           = TaskRoot + "/bucket"
	TaskBucketContentRoot    = TaskBucketRoot + "/*" + Wildcard
	TaskSubmitRoot           = TaskRoot + "/submit"
	TaskCancelRoot           = TaskRoot + "/cancel"
)

const (
	Submit = "submit"
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
	routeGroup.POST(TasksReportQueueRootByIds, h.GetMultiple)
	routeGroup.PUT(TaskRoot, h.Update)
	routeGroup.PATCH(TaskRoot, Transaction, h.Update)
	routeGroup.DELETE(TaskRoot, h.Delete)
	routeGroup.GET(TasksReportQueueRoot, h.Queued)
	routeGroup.GET(TasksReportDashboardRoot, h.Dashboard)
	// Actions
	routeGroup.PUT(TaskSubmitRoot, Transaction, h.Submit)
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
	// Attached
	routeGroup.GET(TaskAttachedRoot, h.GetAttached)
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

// GetMultiple godoc
// @summary Get tasks by a list of IDs.
// @description Get multiple tasks by their IDs.
// @tags tasks
// @produce json
// @success 200 {array} api.Task
// @router /tasks/multiple [post]
// @param ids body []int true "List of Task IDs"
func (h TaskHandler) GetMultiple(ctx *gin.Context) {
	var ids []int
	var tasks []model.Task

	// Parse the body to get the list of IDs
	if err := ctx.ShouldBindJSON(&ids); err != nil {
		h.Respond(ctx, http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Query the database to find all tasks with the given IDs
	db := h.DB(ctx).Preload(clause.Associations)
	result := db.Find(&tasks, ids)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	
}


// List godoc
// @summary List all tasks.
// @description List all tasks.
// @description Filters:
// @description - kind
// @description - createUser
// @description - addon
// @description - name
// @description - locator
// @description - state
// @description - application.id
// @description - application.name
// @description The state=queued is an alias for queued states.
// @tags tasks
// @produce json
// @success 200 {object} []api.Task
// @router /tasks [get]
func (h TaskHandler) List(ctx *gin.Context) {
	resources := []Task{}
	// filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "createUser", Kind: qf.STRING},
			{Field: "kind", Kind: qf.STRING},
			{Field: "addon", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "locator", Kind: qf.STRING},
			{Field: "state", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.STRING},
			{Field: "application.name", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if state, found := filter.Field("state"); found {
		values := qf.Value{}
		for _, v := range state.Value.ByKind(qf.LITERAL, qf.STRING) {
			switch v.Value {
			case "queued":
				values = append(
					values,
					qf.Token{Kind: qf.STRING, Value: tasking.Ready},
					qf.Token{Kind: qf.STRING, Value: tasking.Postponed},
					qf.Token{Kind: qf.STRING, Value: tasking.Pending},
					qf.Token{Kind: qf.STRING, Value: tasking.QuotaBlocked},
					qf.Token{Kind: qf.STRING, Value: tasking.Running})
			default:
				values = append(values, v)
			}
		}
		values = values.Join(qf.OR)
		filter = filter.Revalued("state", values)
	}
	filter = filter.Renamed("application.id", "application__id")
	filter = filter.Renamed("application.name", "application__name")
	filter = filter.Renamed("createUser", "task\\.createUser")
	filter = filter.Renamed("id", "task\\.id")
	filter = filter.Renamed("name", "task\\.name")
	// sort
	sort := Sort{}
	sort.Add("task.id", "id")
	sort.Add("task.createUser", "createUser")
	sort.Add("task.name", "name")
	sort.Add("application__id", "application.id")
	sort.Add("application__name", "application.name")
	err = sort.With(ctx, &model.Task{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Fetch
	db := h.DB(ctx)
	db = db.Model(&model.Task{})
	db = db.Joins("Application")
	db = db.Joins("Report")
	db = sort.Sorted(db)
	db = filter.Where(db)
	var m model.Task
	var list []model.Task
	page := Page{}
	page.With(ctx)
	cursor := Cursor{}
	cursor.With(db, page)
	defer func() {
		cursor.Close()
	}()
	for cursor.Next(&m) {
		if cursor.Error != nil {
			_ = ctx.Error(cursor.Error)
			return
		}
		list = append(list, m)
	}
	err = h.WithCount(ctx, cursor.Count())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		r := Task{}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Queued godoc
// @summary Queued queued task report.
// @description Queued queued task report.
// @description Filters:
// @description - addon
// @tags tasks
// @produce json
// @success 200 {object} []api.TaskQueue
// @router /tasks [get]
func (h TaskHandler) Queued(ctx *gin.Context) {
	r := TaskQueue{}
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "addon", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Table("task")
	db = filter.Where(db)
	type M struct {
		State string
		Count int
	}
	db = db.Select("State", "COUNT(*) Count")
	db = db.Where(
		"State", []string{
			tasking.Ready,
			tasking.Postponed,
			tasking.Pending,
			tasking.QuotaBlocked,
			tasking.Running,
		})
	db = db.Group("State")
	var list []M
	err = db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for _, q := range list {
		r.Total += q.Count
		switch q.State {
		case tasking.Ready:
			r.Ready = q.Count
		case tasking.Postponed:
			r.Postponed = q.Count
		case tasking.Pending:
			r.Pending = q.Count
		case tasking.QuotaBlocked:
			r.QuotaBlocked = q.Count
		case tasking.Running:
			r.Running = q.Count
		}
	}

	h.Respond(ctx, http.StatusOK, r)
}

// Dashboard godoc
// @summary List all task dashboard resources.
// @description List all task dashboard resources.
// @description Filters:
// @description - kind
// @description - createUser
// @description - addon
// @description - name
// @description - locator
// @description - state
// @description - application.id
// @description - application.name
// @tags tasks
// @produce json
// @success 200 {object} []api.TaskDashboard
// @router /tasks [get]
func (h TaskHandler) Dashboard(ctx *gin.Context) {
	resources := []TaskDashboard{}
	// filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "createUser", Kind: qf.STRING},
			{Field: "kind", Kind: qf.STRING},
			{Field: "addon", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "locator", Kind: qf.STRING},
			{Field: "state", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.STRING},
			{Field: "application.name", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("application.id", "application__id")
	filter = filter.Renamed("application.name", "application__name")
	filter = filter.Renamed("createUser", "task\\.createUser")
	filter = filter.Renamed("id", "task\\.id")
	filter = filter.Renamed("name", "task\\.name")
	// sort
	sort := Sort{}
	sort.Add("task.id", "id")
	sort.Add("task.createUser", "createUser")
	sort.Add("task.name", "name")
	sort.Add("application__id", "application.id")
	sort.Add("application__name", "application.name")
	err = sort.With(ctx, &model.Task{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Fetch
	db := h.DB(ctx)
	db = db.Model(&model.Task{})
	db = db.Joins("Application")
	db = db.Joins("Report")
	db = sort.Sorted(db)
	db = filter.Where(db)
	var list []model.Task
	page := Page{}
	page.With(ctx)
	err = db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		r := TaskDashboard{}
		r.With(m)
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
	r := &Task{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	rtx := RichContext(ctx)
	task := &tasking.Task{}
	task.With(r.Model())
	task.CreateUser = h.BaseHandler.CurrentUser(ctx)
	err = rtx.TaskManager.Create(h.DB(ctx), task)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r.With(task.Task)

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
	rtx := RichContext(ctx)
	err := rtx.TaskManager.Delete(h.DB(ctx), id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a task.
// @description Update a task.
// @tags tasks
// @accept json
// @success 200
// @router /tasks/{id} [put]
// @param id path int true "Task ID"
// @param task body Task true "Task data"
func (h TaskHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Task{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &Task{}
	if ctx.Request.Method == http.MethodPatch &&
		ctx.Request.ContentLength > 0 {
		r.With(m)
	}
	err = h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if _, found := ctx.Get(Submit); found {
		r.State = tasking.Ready
	}
	m = r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	rtx := RichContext(ctx)
	task := &tasking.Task{}
	task.With(m)
	err = rtx.TaskManager.Update(h.DB(ctx), task)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// Submit godoc
// @summary Submit a task.
// @description Patch and submit a task.
// @tags tasks
// @accept json
// @success 200
// @router /tasks/{id}/submit [put]
// @param id path int true "Task ID"
// @param task body Task false "Task data (optional)"
func (h TaskHandler) Submit(ctx *gin.Context) {
	ctx.Set(Submit, true)
	ctx.Request.Method = http.MethodPatch
	h.Update(ctx)
}

// Cancel godoc
// @summary Cancel a task.
// @description Cancel a task.
// @tags tasks
// @success 202
// @router /tasks/{id}/cancel [put]
// @param id path int true "Task ID"
func (h TaskHandler) Cancel(ctx *gin.Context) {
	id := h.pk(ctx)
	rtx := RichContext(ctx)
	err := rtx.TaskManager.Cancel(h.DB(ctx), id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusAccepted)
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
		_ = ctx.Error(err)
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
// @success 204
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
	result := db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}

	h.Status(ctx, http.StatusNoContent)
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

// GetAttached godoc
// @summary Get attached files.
// @description Get attached files.
// @description Returns a tarball with attached files.
// @tags tasks
// @produce octet-stream
// @success 200
// @router /tasks/{id}/attached [get]
// @param id path int true "Task ID"
func (h TaskHandler) GetAttached(ctx *gin.Context) {
	m := &model.Task{}
	id := h.pk(ctx)
	db := h.DB(ctx).Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tarWriter := tar.NewWriter(ctx.Writer)
	defer func() {
		tarWriter.Close()
	}()
	r := Task{}
	r.With(m)
	var files []*model.File
	for _, ref := range r.Attached {
		file := &model.File{}
		err = h.DB(ctx).First(file, ref.ID).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		err = tarWriter.AssertFile(file.Path)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		files = append(files, file)
	}
	ctx.Status(http.StatusOK)
	for _, file := range files {
		_ = tarWriter.AddFile(
			file.Path,
			fmt.Sprintf("%.3d-%s", file.ID, file.Name))
	}
}

// TTL time-to-live.
type TTL model.TTL

// TaskPolicy scheduling policies.
type TaskPolicy model.TaskPolicy

// Map unstructured object.
type Map model.Map

// TaskError used in Task.Errors.
type TaskError struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// TaskEvent task event.
type TaskEvent model.TaskEvent

// Attachment file attachment.
type Attachment model.Attachment

// Task REST resource.
type Task struct {
	Resource    `yaml:",inline"`
	Name        string       `json:"name,omitempty" yaml:",omitempty"`
	Kind        string       `json:"kind,omitempty" yaml:",omitempty"`
	Addon       string       `json:"addon,omitempty" yaml:",omitempty"`
	Extensions  []string     `json:"extensions,omitempty" yaml:",omitempty"`
	State       string       `json:"state,omitempty" yaml:",omitempty"`
	Locator     string       `json:"locator,omitempty" yaml:",omitempty"`
	Priority    int          `json:"priority,omitempty" yaml:",omitempty"`
	Policy      TaskPolicy   `json:"policy,omitempty" yaml:",omitempty"`
	TTL         TTL          `json:"ttl,omitempty" yaml:",omitempty"`
	Data        any          `json:"data,omitempty" yaml:",omitempty"`
	Application *Ref         `json:"application,omitempty" yaml:",omitempty"`
	Bucket      *Ref         `json:"bucket,omitempty" yaml:",omitempty"`
	Pod         string       `json:"pod,omitempty" yaml:",omitempty"`
	Retries     int          `json:"retries,omitempty" yaml:",omitempty"`
	Started     *time.Time   `json:"started,omitempty" yaml:",omitempty"`
	Terminated  *time.Time   `json:"terminated,omitempty" yaml:",omitempty"`
	Events      []TaskEvent  `json:"events,omitempty" yaml:",omitempty"`
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
	r.Extensions = m.Extensions
	r.State = m.State
	r.Locator = m.Locator
	r.Priority = m.Priority
	r.Policy = TaskPolicy(m.Policy)
	r.TTL = TTL(m.TTL)
	r.Data = m.Data.Any
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.Pod = m.Pod
	r.Retries = m.Retries
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Events = make([]TaskEvent, 0)
	r.Errors = make([]TaskError, 0)
	r.Attached = make([]Attachment, 0)
	for _, event := range m.Events {
		r.Events = append(r.Events, TaskEvent(event))
	}
	for _, err := range m.Errors {
		r.Errors = append(r.Errors, TaskError(err))
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
	for _, a := range m.Attached {
		r.Attached = append(r.Attached, Attachment(a))
	}
	if Settings.Hub.Task.Preemption.Enabled {
		r.Policy.PreemptEnabled = true
	}
}

// Model builds a model.
func (r *Task) Model() (m *model.Task) {
	m = &model.Task{
		Name:          r.Name,
		Kind:          r.Kind,
		Addon:         r.Addon,
		Extensions:    r.Extensions,
		State:         r.State,
		Locator:       r.Locator,
		Priority:      r.Priority,
		Policy:        model.TaskPolicy(r.Policy),
		TTL:           model.TTL(r.TTL),
		ApplicationID: r.idPtr(r.Application),
	}
	m.ID = r.ID
	m.Data.Any = r.Data
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
	Result    any          `json:"result,omitempty" yaml:",omitempty"`
	TaskID    uint         `json:"task"`
}

// With updates the resource with the model.
func (r *TaskReport) With(m *model.TaskReport) {
	r.Resource.With(&m.Model)
	r.Status = m.Status
	r.Total = m.Total
	r.Completed = m.Completed
	r.TaskID = m.TaskID
	r.Activity = m.Activity
	r.Result = m.Result.Any
	r.Errors = make([]TaskError, 0)
	r.Attached = make([]Attachment, 0)
	for _, err := range m.Errors {
		r.Errors = append(r.Errors, TaskError(err))
	}
	for _, a := range m.Attached {
		r.Attached = append(r.Attached, Attachment(a))
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
		Activity:  r.Activity,
		TaskID:    r.TaskID,
	}
	m.ID = r.ID
	m.Result.Any = r.Result
	for _, err := range r.Errors {
		m.Errors = append(m.Errors, model.TaskError(err))
	}
	for _, at := range r.Attached {
		m.Attached = append(m.Attached, model.Attachment(at))
	}
	return
}

// TaskQueue report.
type TaskQueue struct {
	Total        int `json:"total"`
	Ready        int `json:"ready"`
	Postponed    int `json:"postponed"`
	QuotaBlocked int `json:"quotaBlocked"`
	Pending      int `json:"pending"`
	Running      int `json:"running"`
}

// TaskDashboard report.
type TaskDashboard struct {
	Resource    `yaml:",inline"`
	Name        string     `json:"name,omitempty" yaml:",omitempty"`
	Kind        string     `json:"kind,omitempty" yaml:",omitempty"`
	Addon       string     `json:"addon,omitempty" yaml:",omitempty"`
	State       string     `json:"state,omitempty" yaml:",omitempty"`
	Locator     string     `json:"locator,omitempty" yaml:",omitempty"`
	Application *Ref       `json:"application,omitempty" yaml:",omitempty"`
	Started     *time.Time `json:"started,omitempty" yaml:",omitempty"`
	Terminated  *time.Time `json:"terminated,omitempty" yaml:",omitempty"`
	Errors      int        `json:"errors,omitempty" yaml:",omitempty"`
}

func (r *TaskDashboard) With(m *model.Task) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.State = m.State
	r.Locator = m.Locator
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Errors = len(m.Errors)
	if m.Report != nil {
		r.Errors += len(m.Report.Errors)
	}
}
