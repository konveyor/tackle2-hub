package api

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

//
// Routes
const (
	TasksRoot      = "/tasks"
	TaskRoot       = TasksRoot + "/:" + ID
	TaskReportRoot = TaskRoot + "/report"
	AddonTasksRoot = AddonRoot + "/tasks"
)

const (
	LocatorParam = "locator"
)

//
// TaskHandler handles task routes.
type TaskHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h TaskHandler) AddRoutes(e *gin.Engine) {
	e.GET(TasksRoot, h.List)
	e.GET(TasksRoot+"/", h.List)
	e.POST(TasksRoot, h.Create)
	e.GET(TaskRoot, h.Get)
	e.PUT(TaskRoot, h.Update)
	e.POST(TaskReportRoot, h.CreateReport)
	e.PUT(TaskReportRoot, h.UpdateReport)
	e.POST(AddonTasksRoot, h.AddonCreate)
	e.GET(AddonTasksRoot, h.AddonList)
	e.DELETE(TaskRoot, h.Delete)
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
	id := ctx.Param(ID)
	db := h.DB.Preload("Report")
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
	db = db.Preload("Report")
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
	task := Task{}
	err := ctx.BindJSON(&task)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}

	m := task.Model()
	m.Reset()
	result := h.DB.Create(&m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	task.With(m)

	ctx.JSON(http.StatusCreated, task)
}

// Delete godoc
// @summary Delete a task.
// @description Delete a task.
// @tags delete
// @success 204
// @router /tasks/{id} [delete]
// @param id path string true "Task ID"
func (h TaskHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	task := &model.Task{}
	result := h.DB.First(task, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}
	if task.Job != "" {
		job := &batch.Job{}
		job.Namespace = path.Dir(task.Job)
		job.Name = path.Base(task.Job)
		err := h.Client.Delete(
			context.TODO(),
			job)
		if err != nil {
			h.deleteFailed(ctx, result.Error)
		}
	}
	result = h.DB.Delete(task, id)
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
	id := ctx.Param(ID)
	updates := &Task{}
	err := ctx.BindJSON(updates)
	if err != nil {
		return
	}
	m := updates.Model()
	result := h.DB.Model(&Task{}).Where("id", id).Omit("id").Updates(m)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// CreateReport godoc
// @summary Create a task report.
// @description Update a task report.
// @tags update
// @accept json
// @produce json
// @success 201 {object} api.TaskReport
// @router /tasks/{id}/report [post]
// @param id path string true "TaskReport ID"
// @param task body api.TaskReport true "TaskReport data"
func (h TaskHandler) CreateReport(ctx *gin.Context) {
	id := ctx.Param(ID)
	report := &TaskReport{}
	err := ctx.BindJSON(report)
	if err != nil {
		return
	}
	task, _ := strconv.Atoi(id)
	report.TaskID = uint(task)
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
// @param id path string true "TaskReport ID"
// @param task body api.TaskReport true "TaskReport data"
func (h TaskHandler) UpdateReport(ctx *gin.Context) {
	id := ctx.Param(ID)
	report := &TaskReport{}
	err := ctx.BindJSON(report)
	if err != nil {
		return
	}
	task, _ := strconv.Atoi(id)
	report.TaskID = uint(task)
	m := report.Model()
	db := h.DB.Model(m)
	db = db.Where("taskid", task)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
	}
	report.With(m)

	ctx.JSON(http.StatusOK, report)
}

// AddonCreate godoc
// @summary Create an addon task.
// @description Create an addon task.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.Task
// @router /addons/:name/tasks [post]
// @param task body api.Task true "Task data"
func (h TaskHandler) AddonCreate(ctx *gin.Context) {
	name := ctx.Param(Name)
	addon := &crd.Addon{}
	err := h.Client.Get(
		context.TODO(),
		client.ObjectKey{
			Namespace: Settings.Hub.Namespace,
			Name:      name,
		},
		addon)
	if err != nil {
		if errors.IsNotFound(err) {
			ctx.Status(http.StatusNotFound)
			return
		}
	}
	task := Task{}
	task.Name = addon.Name
	task.Addon = addon.Name
	task.Image = addon.Spec.Image
	err = ctx.BindJSON(&task.Data)
	if err != nil {
		return
	}
	m := task.Model()
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	task.With(m)

	ctx.JSON(http.StatusCreated, task)
}

// AddonList godoc
// @summary List all tasks associated to an addon.
// @description List all tasks associated to an addon.
// @tags get
// @produce json
// @success 200 {object} []api.Task
// @router /addons/{name}/tasks [get]
func (h TaskHandler) AddonList(ctx *gin.Context) {
	var list []model.Task
	name := ctx.Param(Name)
	db := h.DB.Where("addon", name)
	locator := ctx.Query(LocatorParam)
	if locator != "" {
		db = db.Where("locator", locator)
	}
	db = db.Preload("Report")
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

//
// AddonTask REST resource.
type AddonTask struct {
	Resource
	Name     string      `json:"name"`
	Locator  string      `json:"locator"`
	Isolated bool        `json:"isolated,omitempty"`
	Data     interface{} `json:"data" swaggertype:"object"`
}

//
// Task REST resource.
type Task struct {
	Resource
	Name       string      `json:"name"`
	Locator    string      `json:"locator"`
	Isolated   bool        `json:"isolated,omitempty"`
	Data       interface{} `json:"data" swaggertype:"object"`
	Addon      string      `json:"addon"`
	Image      string      `json:"image"`
	Started    *time.Time  `json:"started"`
	Terminated *time.Time  `json:"terminated"`
	Status     string      `json:"status"`
	Error      string      `json:"error"`
	Job        string      `json:"job"`
	Report     *TaskReport `json:"report"`
}

//
// With updates the resource with the model.
func (r *Task) With(m *model.Task) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Image = m.Image
	r.Addon = m.Addon
	r.Locator = m.Locator
	r.Isolated = m.Isolated
	r.Started = m.Started
	r.Terminated = m.Terminated
	r.Status = m.Status
	r.Error = m.Error
	r.Job = m.Job
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
		Name:     r.Name,
		Addon:    r.Addon,
		Locator:  r.Locator,
		Isolated: r.Isolated,
	}
	m.Data, _ = json.Marshal(r.Data)
	m.ID = r.ID
	return
}

//
// TaskReport REST resource.
type TaskReport struct {
	Resource
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	Total     int      `json:"total"`
	Completed int      `json:"completed"`
	Activity  []string `json:"activity"`
	TaskID    uint     `json:"task"`
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
	_ = json.Unmarshal(m.Activity, &r.Activity)
}

//
// Model builds a model.
func (r *TaskReport) Model() (m *model.TaskReport) {
	m = &model.TaskReport{
		Status:    r.Status,
		Error:     r.Error,
		Total:     r.Total,
		Completed: r.Completed,
		TaskID:    r.TaskID,
	}
	if r.Activity == nil {
		r.Activity = []string{}
	}
	_ = json.Unmarshal(m.Activity, &r.Activity)
	m.Activity, _ = json.Marshal(r.Activity)
	m.ID = r.ID

	return
}
