package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	routeGroup.PUT(TaskGroupSubmitRoot, h.Submit)
	routeGroup.GET(TaskGroupBucketRoot, h.Content)
	routeGroup.POST(TaskGroupBucketRoot, h.Upload)
	routeGroup.PUT(TaskGroupBucketRoot, h.Upload)
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
	group := &TaskGroup{}
	err := ctx.BindJSON(group)
	if err != nil {
		h.createFailed(ctx, err)
		return
	}
	m := group.Model()
	result := h.DB.Create(&m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}

	group.With(m)

	ctx.JSON(http.StatusCreated, group)
}

// Update godoc
// @summary Update a task group.
// @description Update a task group.
// @tags update
// @accept json
// @success 204
// @router /tasks/{id} [put]
// @param id path string true "Task ID"
// @param task body Task true "Task data"
func (h TaskGroupHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &TaskGroup{}
	err := ctx.BindJSON(r)
	if err != nil {
		return
	}
	current := &model.TaskGroup{}
	result := h.DB.First(current, id)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	updated := r.Model()
	updated.ID = current.ID
	updated.Bucket = current.Bucket
	err = h.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			db := tx.Model(updated)
			db = db.Omit(clause.Associations)
			result := db.Updates(h.fields(updated))
			if result.Error != nil {
				err = result.Error
				return
			}
			wanted := []uint{}
			for i := range updated.Tasks {
				m := &updated.Tasks[i]
				m.TaskGroupID = &id
				if m.ID == 0 {
					result := tx.Create(m)
					if result.Error != nil {
						err = result.Error
						return
					}
				} else {
					db := tx.Model(m)
					db = db.Where("status", tasking.Created)
					result := db.Save(m)
					if result.Error != nil {
						err = result.Error
						return
					}
				}
				wanted = append(wanted, m.ID)
			}
			db = tx.Where("id NOT IN ?", wanted)
			db = db.Where("taskgroupid", id)
			var unwanted []model.Task
			result = db.Find(&unwanted)
			if result.Error != nil {
				err = result.Error
				return
			}
			for i := range unwanted {
				task := &unwanted[i]
				result = db.Delete(task)
				if result.Error != nil {
					err = result.Error
					return
				}
			}
			return
		})
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
	err := h.DB.Transaction(
		func(tx *gorm.DB) (err error) {
			db := tx.Preload(clause.Associations)
			result := db.First(m, id)
			if result.Error != nil {
				err = result.Error
				return
			}
			result = db.Delete(m)
			if result.Error != nil {
				err = result.Error
				return
			}
			for i := range m.Tasks {
				m := &m.Tasks[i]
				result := tx.Delete(m)
				if result.Error != nil {
					err = result.Error
					return
				}
			}

			return
		})
	if err != nil {
		h.updateFailed(ctx, err)
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
// @router /taskgroups/{id}/submit [post]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) Submit(ctx *gin.Context) {
	id := h.pk(ctx)
	result := h.DB.First(&model.TaskGroup{}, id)
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	db := h.DB.Model(&model.Task{})
	db = db.Where("taskgroupid", id)
	db = db.Where("status", tasking.Created)
	result = db.Updates(
		model.Map{
			"status": tasking.Ready,
		})
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}
	if result.RowsAffected > 0 {
		ctx.Status(http.StatusAccepted)
		return
	}

	ctx.Status(http.StatusOK)
}

// Content godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @tags get
// @produce octet-stream
// @success 200
// @router /taskgroups/{id}/bucket/{wildcard} [get]
// @param id path string true "TaskGroup ID"
func (h TaskGroupHandler) Content(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.TaskGroup{}
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
func (h TaskGroupHandler) Upload(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}

	h.upload(ctx, &m.BucketOwner)
}

//
// TaskGroup REST resource.
type TaskGroup struct {
	Resource
	Name   string      `json:"name"`
	Addon  string      `json:"addon"`
	Data   interface{} `json:"data" swaggertype:"object"`
	Bucket string      `json:"bucket"`
	Purged bool        `json:"purged,omitempty"`
	Tasks  []Task      `json:"tasks"`
}

//
// With updates the resource with the model.
func (r *TaskGroup) With(m *model.TaskGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Addon = m.Addon
	r.Bucket = m.Bucket
	r.Purged = m.Purged
	r.Tasks = []Task{}
	for _, task := range m.Tasks {
		member := Task{}
		member.With(&task)
		r.Tasks = append(
			r.Tasks,
			member)
	}
	_ = json.Unmarshal(m.Data, &r.Data)
}

//
// Model builds a model.
func (r *TaskGroup) Model() (m *model.TaskGroup) {
	m = &model.TaskGroup{
		Name:   r.Name,
		Addon:  r.Addon,
		Purged: r.Purged,
	}
	for _, task := range r.Tasks {
		member := *task.Model()
		member.Status = tasking.Created
		m.Tasks = append(
			m.Tasks,
			member)
	}
	m.Data, _ = json.Marshal(r.Data)
	m.ID = r.ID
	return
}
