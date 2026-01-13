package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/internal/api/filter"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/tar"
	"gorm.io/gorm/clause"
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
	routeGroup.GET(api.TasksRoute, h.List)
	routeGroup.GET(api.TasksRoute+"/", h.List)
	routeGroup.POST(api.TasksRoute, h.Create)
	routeGroup.GET(api.TaskRoute, h.Get)
	routeGroup.PUT(api.TaskRoute, h.Update)
	routeGroup.PATCH(api.TaskRoute, Transaction, h.Update)
	routeGroup.DELETE(api.TaskRoute, h.Delete)
	routeGroup.GET(api.TasksReportQueueRoute, h.Queued)
	routeGroup.GET(api.TasksReportDashboardRoute, h.Dashboard)
	// Actions
	routeGroup.PUT(api.TaskSubmitRoute, Transaction, h.Submit)
	routeGroup.PUT(api.TaskCancelRoute, h.Cancel)
	routeGroup.PUT(api.TasksCancelRoute, h.BulkCancel)
	// Bucket
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tasks.bucket"))
	routeGroup.GET(api.TaskBucketRoute, h.BucketGet)
	routeGroup.GET(api.TaskBucketContentRoute, h.BucketGet)
	routeGroup.POST(api.TaskBucketContentRoute, h.BucketPut)
	routeGroup.PUT(api.TaskBucketContentRoute, h.BucketPut)
	routeGroup.DELETE(api.TaskBucketContentRoute, h.BucketDelete)
	// Report
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tasks.report"))
	routeGroup.POST(api.TaskReportRoute, h.CreateReport)
	routeGroup.PUT(api.TaskReportRoute, h.UpdateReport)
	routeGroup.DELETE(api.TaskReportRoute, h.DeleteReport)
	// Attached
	routeGroup.GET(api.TaskAttachedRoute, h.GetAttached)
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
	err := db.First(task, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := Task{}
	r.With(task)
	q := ctx.Query("merged")
	if b, _ := strconv.ParseBool(q); b {
		err := r.InjectFiles(h.DB(ctx))
		if err != nil {
			_ = ctx.Error(err)
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
// @description - createUser
// @description - addon
// @description - name
// @description - locator
// @description - state
// @description - application.id
// @description - application.name
// @description - platform.id
// @description - platform.name
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
			{Field: "platform.id", Kind: qf.STRING},
			{Field: "platform.name", Kind: qf.STRING},
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
					qf.Token{Kind: qf.STRING, Value: task.Ready},
					qf.Token{Kind: qf.STRING, Value: task.Postponed},
					qf.Token{Kind: qf.STRING, Value: task.Pending},
					qf.Token{Kind: qf.STRING, Value: task.QuotaBlocked},
					qf.Token{Kind: qf.STRING, Value: task.Running})
			default:
				values = append(values, v)
			}
		}
		values = values.Join(qf.OR)
		filter = filter.Revalued("state", values)
	}
	filter = filter.Renamed("application.id", "application__id")
	filter = filter.Renamed("application.name", "application__name")
	filter = filter.Renamed("platform.id", "platform__id")
	filter = filter.Renamed("platform.name", "platform__name")
	filter = filter.Renamed("createUser", "task\\.createUser")
	filter = filter.Renamed("id", "task\\.id")
	filter = filter.Renamed("name", "task\\.name")
	filter = filter.Renamed("kind", "task\\.kind")
	// sort
	sort := Sort{}
	sort.Add("task.id", "id")
	sort.Add("task.createUser", "createUser")
	sort.Add("task.name", "name")
	sort.Add("task.kind", "kind")
	sort.Add("application__id", "application.id")
	sort.Add("application__name", "application.name")
	sort.Add("platform__id", "platform.id")
	sort.Add("platform__name", "platform.name")
	err = sort.With(ctx, &model.Task{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Fetch
	db := h.DB(ctx)
	db = db.Model(&model.Task{})
	db = db.Joins("Application")
	db = db.Joins("Platform")
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
			task.Ready,
			task.Postponed,
			task.Pending,
			task.QuotaBlocked,
			task.Running,
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
		case task.Ready:
			r.Ready = q.Count
		case task.Postponed:
			r.Postponed = q.Count
		case task.Pending:
			r.Pending = q.Count
		case task.QuotaBlocked:
			r.QuotaBlocked = q.Count
		case task.Running:
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
// @description - platform.id
// @description - platform.name
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
			{Field: "platform.id", Kind: qf.STRING},
			{Field: "platform.name", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("application.id", "application__id")
	filter = filter.Renamed("application.name", "application__name")
	filter = filter.Renamed("platform.id", "platform__id")
	filter = filter.Renamed("platform.name", "platform__name")
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
	sort.Add("platform__id", "platform.id")
	sort.Add("platform__name", "platform.name")
	err = sort.With(ctx, &model.Task{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Fetch
	db := h.DB(ctx)
	db = db.Model(&model.Task{})
	db = db.Joins("Application")
	db = db.Joins("Platform")
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
	m := &model.Task{}
	r.Patch(m)
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	task := task.NewTask(m)
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
		r.State = task.Ready
	}
	r.Patch(m)
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	rtx := RichContext(ctx)
	task := task.NewTask(m)
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

// BulkCancel godoc
// @summary Cancel tasks matched by the filter.
// @description Cancel tasks matched by the filter.
// @description Caution: an empty filter matches all tasks.
// @description Filters:
// @description - id
// @description - name
// @description - locator
// @description - kind
// @description - addon
// @description - state
// @description - application.id
// @tags tasks
// @success 202
// @router /tasks/cancel/list [put]
// @param tasks body []uint true "List of Task IDs"
func (h TaskHandler) BulkCancel(ctx *gin.Context) {
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "id", Kind: qf.LITERAL},
			{Field: "name", Kind: qf.STRING},
			{Field: "locator", Kind: qf.STRING},
			{Field: "kind", Kind: qf.STRING},
			{Field: "addon", Kind: qf.STRING},
			{Field: "state", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.STRING},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("application.id", "applicationId")
	db := h.DB(ctx)
	db = db.Model(&model.Task{})
	db = filter.Where(db)
	matched := []*model.Task{}
	err = db.Find(&matched).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx)
	rtx := RichContext(ctx)
	for _, m := range matched {
		err := rtx.TaskManager.Cancel(db, m.ID)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
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
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
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
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
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
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
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
	r := &TaskReport{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := &model.TaskReport{}
	r.Patch(m)
	m.TaskID = id
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	err = h.DB(ctx).Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
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
	r := &TaskReport{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := &model.TaskReport{}
	r.Patch(m)
	m.TaskID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Where("taskid", id)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
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
	err := db.Delete(&model.TaskReport{}).Error
	if err != nil {
		_ = ctx.Error(err)
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
type TTL = resource.TTL

// TaskPolicy scheduling policies.
type TaskPolicy = model.TaskPolicy

// TaskError used in Task.Errors.
type TaskError = resource.TaskError

// TaskEvent task event.
type TaskEvent model.TaskEvent

// Attachment file attachment.
type Attachment model.Attachment

// Task REST resource.
type Task = resource.Task

// TaskReport REST resource.
type TaskReport = resource.TaskReport

// TaskQueue report.
type TaskQueue = resource.TaskQueue

// TaskDashboard report.
type TaskDashboard = resource.TaskDashboard
