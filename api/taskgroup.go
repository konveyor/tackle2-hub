package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
)

//
// Routes
const (
	TaskGroupsRoot      = "/taskgroups"
	TaskGroupRoot       = TaskGroupsRoot + "/:" + ID
	TaskGroupBucketRoot = TaskGroupRoot + "/bucket/*" + Wildcard
	TaskGroupSubmitRoot = TaskGroupRoot + "/submit"
)

//
// TaskGroupHandler handles task group routes.
type TaskGroupHandler struct {
	BaseHandler
	BucketHandler
}

//
// AddRoutes adds routes.
func (h TaskGroupHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, "tasks"))
	routeGroup.GET(TaskGroupsRoot, h.List)
	routeGroup.GET(TaskGroupsRoot+"/", h.List)
	routeGroup.POST(TaskGroupsRoot, h.Create)
	routeGroup.PUT(TaskGroupRoot, h.Update)
	routeGroup.GET(TaskGroupRoot, h.Get)
	routeGroup.PUT(TaskGroupSubmitRoot, h.Submit, h.Update)
	routeGroup.GET(TaskGroupBucketRoot, h.BucketGet)
	routeGroup.POST(TaskGroupBucketRoot, h.BucketUpload)
	routeGroup.PUT(TaskGroupBucketRoot, h.BucketUpload)
	routeGroup.DELETE(TaskGroupBucketRoot, h.BucketDelete)
	routeGroup.DELETE(TaskGroupRoot, h.Delete)
}

// Get godoc
// @summary Get a task group by ID.
// @description Get a task group by ID.
// @tags get
// @produce json
// @success 200 {object} api.TaskGroup
// @router /taskgroups/{id} [get]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) Get(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	db := h.DB.Preload(clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := TaskGroup{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all task groups.
// @description List all task groups.
// @tags get
// @produce json
// @success 200 {object} []api.TaskGroup
// @router /taskgroups [get]
func (h TaskGroupHandler) List(ctx *gin.Context) {
	var list []model.TaskGroup
	db := h.DB.Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []TaskGroup{}
	for i := range list {
		r := TaskGroup{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a task group.
// @description Create a task group.
// @tags create
// @accept json
// @produce json
// @success 201 {object} api.TaskGroup
// @router /taskgroups [post]
// @param taskgroup body api.TaskGroup true "TaskGroup data"
func (h TaskGroupHandler) Create(ctx *gin.Context) {
	r := &TaskGroup{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	db := h.DB
	m := r.Model()
	switch r.State {
	case "":
		m.State = tasking.Created
		fallthrough
	case tasking.Created:
		db = h.DB.Omit(clause.Associations)
	case tasking.Ready:
		err := m.Propagate()
		if err != nil {
			return
		}
	default:
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "state must be ('''|Created|Ready)",
			})
		return
	}
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := db.Create(&m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}

	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Update godoc
// @summary Update a task group.
// @description Update a task group.
// @tags update
// @accept json
// @success 204
// @router /taskgroups/{id} [put]
// @param id path string true "Task ID"
// @param task body Task true "Task data"
func (h TaskGroupHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	updated := &TaskGroup{}
	err := ctx.BindJSON(updated)
	if err != nil {
		return
	}
	current := &model.TaskGroup{}
	err = h.DB.First(current, id).Error
	if err != nil {
		h.getFailed(ctx, err)
		return
	}
	m := updated.Model()
	m.ID = current.ID
	m.Bucket = current.Bucket
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB.Model(m)
	switch updated.State {
	case "", tasking.Created:
		db = db.Omit(clause.Associations)
	case tasking.Ready:
		err := m.Propagate()
		if err != nil {
			return
		}
	default:
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "state must be (Created|Ready)",
			})
		return
	}
	db = db.Omit("Bucket")
	db = db.Where("state IN ?", []string{"", tasking.Created})
	err = db.Updates(h.fields(m)).Error
	if err != nil {
		h.updateFailed(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Delete godoc
// @summary Delete a task group.
// @description Delete a task group.
// @tags delete
// @success 204
// @router /taskgroups/{id} [delete]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) Delete(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	db := h.DB.Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		h.deleteFailed(ctx, err)
		return
	}
	for _, task := range m.Tasks {
		if task.Pod != "" {
			rt := tasking.Task{Task: &task}
			err := rt.Delete(h.Client)
			if err != nil {
				if !k8serr.IsNotFound(err) {
					h.deleteFailed(ctx, err)
					return
				}
			}
		}
		db := h.DB.Select(clause.Associations)
		err = db.Delete(task).Error
		if err != nil {
			h.deleteFailed(ctx, err)
			return
		}
	}
	db = h.DB.Select(clause.Associations)
	err = db.Delete(m).Error
	if err != nil {
		h.deleteFailed(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Submit godoc
// @summary Submit a task group.
// @description Submit a task group.
// @tags update
// @accept json
// @success 202
// @router /taskgroups/{id}/submit [put]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) Submit(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &TaskGroup{}
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

// BucketGet godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @tags get
// @produce octet-stream
// @success 200
// @router /taskgroups/{id}/bucket/{wildcard} [get]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) BucketGet(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TaskGroup{}
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	h.content(ctx, &m.BucketOwner)
}

// BucketUpload godoc
// @summary Upload bucket content by ID and path.
// @description Upload bucket content by ID and path.
// @tags post
// @produce json
// @success 204
// @router /taskgroups/{id}/bucket/{wildcard} [post]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) BucketUpload(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	h.upload(ctx, &m.BucketOwner)
}

// BucketDelete godoc
// @summary Delete bucket content by ID and path.
// @description Delete bucket content by ID and path.
// @tags delete
// @produce json
// @success 204
// @router /taskgroups/{id}/bucket/{wildcard} [delete]
// @param id path string true "Task ID"
func (h TaskGroupHandler) BucketDelete(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	h.delete(ctx, &m.BucketOwner)
}

//
// TaskGroup REST resource.
type TaskGroup struct {
	Resource
	Name   string      `json:"name"`
	Addon  string      `json:"addon"`
	Data   interface{} `json:"data" swaggertype:"object" binding:"required"`
	Bucket string      `json:"bucket,omitempty"`
	State  string      `json:"state"`
	Tasks  []Task      `json:"tasks"`
}

//
// With updates the resource with the model.
func (r *TaskGroup) With(m *model.TaskGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Addon = m.Addon
	r.State = m.State
	r.Bucket = m.Bucket
	r.Tasks = []Task{}
	_ = json.Unmarshal(m.Data, &r.Data)
	switch m.State {
	case "", tasking.Created:
		_ = json.Unmarshal(m.List, &r.Tasks)
	default:
		for _, task := range m.Tasks {
			member := Task{}
			member.With(&task)
			r.Tasks = append(
				r.Tasks,
				member)
		}
	}
}

//
// Model builds a model.
func (r *TaskGroup) Model() (m *model.TaskGroup) {
	m = &model.TaskGroup{
		Name:  r.Name,
		Addon: r.Addon,
		State: r.State,
	}
	m.ID = r.ID
	m.Bucket = r.Bucket
	m.Data, _ = json.Marshal(r.Data)
	m.List, _ = json.Marshal(r.Tasks)
	for _, task := range r.Tasks {
		member := task.Model()
		m.Tasks = append(
			m.Tasks,
			*member)
	}
	return
}
