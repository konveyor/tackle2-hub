package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/tar"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/strings/slices"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Routes
const (
	TasksRoot             = "/tasks"
	TaskRoot              = TasksRoot + "/:" + ID
	TaskReportRoot        = TaskRoot + "/report"
	TaskAttachedRoot      = TaskRoot + "/attached"
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

// List godoc
// @summary List all tasks.
// @description List all tasks.
// @description Filters:
// @description - kind
// @description - addon
// @description - name
// @description - locator
// @description - state
// @description - application.id
// @tags tasks
// @produce json
// @success 200 {object} []api.Task
// @router /tasks [get]
func (h TaskHandler) List(ctx *gin.Context) {
	resources := []Task{}
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "kind", Kind: qf.STRING},
			{Field: "addon", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "locator", Kind: qf.STRING},
			{Field: "state", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	sort := Sort{}
	err = sort.With(ctx, &model.Issue{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Model(&model.Task{})
	db = db.Preload(clause.Associations)
	db = sort.Sorted(db)
	filter = filter.Renamed("application.id", "applicationId")
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
	err = h.findRefs(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	task := &tasking.Task{}
	task.With(r.Model())
	task.CreateUser = h.BaseHandler.CurrentUser(ctx)
	rtx := WithContext(ctx)
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
	rtx := WithContext(ctx)
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
	r.ID = id
	rtx := WithContext(ctx)
	task := &tasking.Task{}
	task.With(r.Model())
	task.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	err = rtx.TaskManager.Update(h.DB(ctx), task)
	if err != nil {
		_ = ctx.Error(err)
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
	err := h.findRefs(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
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
	err = h.modBody(ctx, r, mod)
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
	rtx := WithContext(ctx)
	err := rtx.TaskManager.Cancel(h.DB(ctx), id)
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

// findRefs find referenced resources.
// - addon
// - extensions
// - kind
// - priority
// The priority is defaulted to the kind as needed.
func (h *TaskHandler) findRefs(ctx *gin.Context, r *Task) (err error) {
	client := h.Client(ctx)
	if r.Addon != "" {
		addon := &crd.Addon{}
		name := r.Addon
		err = client.Get(
			context.TODO(),
			k8sclient.ObjectKey{
				Name:      name,
				Namespace: Settings.Hub.Namespace,
			},
			addon)
		if err != nil {
			if k8serr.IsNotFound(err) {
				err = &BadRequestError{
					Reason: "Addon: " + name + " not found",
				}
			}
			return
		}
	}
	for _, name := range r.Extensions {
		ext := &crd.Extension{}
		err = client.Get(
			context.TODO(),
			k8sclient.ObjectKey{
				Name:      name,
				Namespace: Settings.Hub.Namespace,
			},
			ext)
		if err != nil {
			if k8serr.IsNotFound(err) {
				err = &BadRequestError{
					Reason: "Extension: " + name + " not found",
				}
			}
			return
		}
	}
	if r.Kind != "" {
		kind := &crd.Task{}
		name := r.Kind
		err = client.Get(
			context.TODO(),
			k8sclient.ObjectKey{
				Name:      name,
				Namespace: Settings.Hub.Namespace,
			},
			kind)
		if err != nil {
			if k8serr.IsNotFound(err) {
				err = &BadRequestError{
					Reason: "Task: " + name + " not found",
				}
			}
			return
		}
		if r.Priority == 0 {
			r.Priority = kind.Spec.Priority
		}
	}
	return
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
	Actions     []string     `json:"actions,omitempty" yaml:",omitempty"`
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
