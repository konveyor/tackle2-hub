package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
)

// Routes
const (
	TaskGroupsRoot             = "/taskgroups"
	TaskGroupRoot              = TaskGroupsRoot + "/:" + ID
	TaskGroupBucketRoot        = TaskGroupRoot + "/bucket"
	TaskGroupBucketContentRoot = TaskGroupBucketRoot + "/*" + Wildcard
	TaskGroupSubmitRoot        = TaskGroupRoot + "/submit"
)

// TaskGroupHandler handles task group routes.
type TaskGroupHandler struct {
	BucketOwner
}

// AddRoutes adds routes.
func (h TaskGroupHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tasks"), Transaction)
	routeGroup.GET(TaskGroupsRoot, h.List)
	routeGroup.GET(TaskGroupsRoot+"/", h.List)
	routeGroup.POST(TaskGroupsRoot, h.Create)
	routeGroup.PUT(TaskGroupRoot, h.Update)
	routeGroup.GET(TaskGroupRoot, h.Get)
	routeGroup.PUT(TaskGroupSubmitRoot, h.Submit, h.Update)
	routeGroup.DELETE(TaskGroupRoot, h.Delete)
	// Bucket
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tasks.bucket"))
	routeGroup.GET(TaskGroupBucketRoot, h.BucketGet)
	routeGroup.GET(TaskGroupBucketContentRoot, h.BucketGet)
	routeGroup.POST(TaskGroupBucketContentRoot, h.BucketPut)
	routeGroup.PUT(TaskGroupBucketContentRoot, h.BucketPut)
	routeGroup.DELETE(TaskGroupBucketContentRoot, h.BucketDelete)
}

// Get godoc
// @summary Get a task group by ID.
// @description Get a task group by ID.
// @tags taskgroups
// @produce json
// @success 200 {object} api.TaskGroup
// @router /taskgroups/{id} [get]
// @param id path int true "TaskGroup ID"
func (h TaskGroupHandler) Get(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	db := h.DB(ctx).Preload(clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := TaskGroup{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all task groups.
// @description List all task groups.
// @tags taskgroups
// @produce json
// @success 200 {object} []api.TaskGroup
// @router /taskgroups [get]
func (h TaskGroupHandler) List(ctx *gin.Context) {
	var list []model.TaskGroup
	db := h.DB(ctx).Preload(clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []TaskGroup{}
	for i := range list {
		r := TaskGroup{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a task group.
// @description Create a task group.
// @tags taskgroups
// @accept json
// @produce json
// @success 201 {object} api.TaskGroup
// @router /taskgroups [post]
// @param taskgroup body api.TaskGroup true "TaskGroup data"
func (h TaskGroupHandler) Create(ctx *gin.Context) {
	r := &TaskGroup{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	m := r.Model()
	switch r.State {
	case "":
		m.State = tasking.Created
		fallthrough
	case tasking.Created:
		db = h.DB(ctx).Omit(clause.Associations)
	case tasking.Ready:
		err := m.Propagate()
		if err != nil {
			return
		}
	default:
		h.Respond(ctx,
			http.StatusBadRequest,
			gin.H{
				"error": "state must be ('''|Created|Ready)",
			})
		return
	}
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := db.Create(&m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Update godoc
// @summary Update a task group.
// @description Update a task group.
// @tags taskgroups
// @accept json
// @success 204
// @router /taskgroups/{id} [put]
// @param id path int true "Task ID"
// @param task body TaskGroup true "Task data"
func (h TaskGroupHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	updated := &TaskGroup{}
	err := h.Bind(ctx, updated)
	if err != nil {
		return
	}
	current := &model.TaskGroup{}
	err = h.DB(ctx).First(current, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := updated.Model()
	m.ID = current.ID
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)

	omit := []string{"BucketID", "Bucket"}
	switch updated.State {
	case "", tasking.Created:
		omit = append(omit, clause.Associations)
	case tasking.Ready:
		err := m.Propagate()
		if err != nil {
			return
		}
	default:
		h.Respond(ctx,
			http.StatusBadRequest,
			gin.H{
				"error": "state must be (Created|Ready)",
			})
		return
	}
	db = db.Omit(omit...)
	db = db.Where("state IN ?", []string{"", tasking.Created})
	err = db.Updates(h.fields(m)).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Delete godoc
// @summary Delete a task group.
// @description Delete a task group.
// @tags taskgroups
// @success 204
// @router /taskgroups/{id} [delete]
// @param id path int true "TaskGroup ID"
func (h TaskGroupHandler) Delete(ctx *gin.Context) {
	m := &model.TaskGroup{}
	id := h.pk(ctx)
	db := h.DB(ctx).Preload(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for _, task := range m.Tasks {
		if task.Pod != "" {
			rt := tasking.Task{Task: &task}
			err := rt.Delete(h.Client(ctx))
			if err != nil {
				if !k8serr.IsNotFound(err) {
					_ = ctx.Error(err)
					return
				}
			}
		}
		db := h.DB(ctx).Select(clause.Associations)
		err = db.Delete(task).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}
	db = h.DB(ctx).Select(clause.Associations)
	err = db.Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Submit godoc
// @summary Submit a task group.
// @description Submit a task group.
// @tags taskgroups
// @accept json
// @success 204
// @router /taskgroups/{id}/submit [put]
// @param id path int true "TaskGroup ID"
// @param taskgroup body TaskGroup false "TaskGroup data (optional)"
func (h TaskGroupHandler) Submit(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &TaskGroup{}
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

// BucketGet godoc
// @summary Get bucket content by ID and path.
// @description Get bucket content by ID and path.
// @description Returns index.html for directories when Accept=text/html else a tarball.
// @description ?filter=glob supports directory content filtering.
// @tags taskgroups
// @produce octet-stream
// @success 200
// @router /taskgroups/{id}/bucket/{wildcard} [get]
// @param id path int true "TaskGroup ID"
// @param wildcard path string true "Content path"
// @param filter query string false "Filter"
func (h TaskGroupHandler) BucketGet(ctx *gin.Context) {
	m := &model.TaskGroup{}
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
// @tags taskgroups
// @produce json
// @success 204
// @router /taskgroups/{id}/bucket/{wildcard} [post]
// @param id path int true "TaskGroup ID"
// @param wildcard path string true "Content path"
func (h TaskGroupHandler) BucketPut(ctx *gin.Context) {
	m := &model.TaskGroup{}
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
// @tags taskgroups
// @produce json
// @success 204
// @router /taskgroups/{id}/bucket/{wildcard} [delete]
// @param id path int true "Task ID"
// @param wildcard path string true "Content path"
func (h TaskGroupHandler) BucketDelete(ctx *gin.Context) {
	m := &model.TaskGroup{}
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

// TaskGroup REST resource.
type TaskGroup struct {
	Resource   `yaml:",inline"`
	Name       string      `json:"name"`
	Kind       string      `json:"kind,omitempty" yaml:",omitempty"`
	Addon      string      `json:"addon,omitempty" yaml:",omitempty"`
	Extensions []string    `json:"extensions,omitempty" yaml:",omitempty"`
	Data       interface{} `json:"data" swaggertype:"object" binding:"required"`
	Bucket     *Ref        `json:"bucket,omitempty"`
	State      string      `json:"state"`
	Tasks      []Task      `json:"tasks"`
}

// With updates the resource with the model.
func (r *TaskGroup) With(m *model.TaskGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.State = m.State
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.Tasks = []Task{}
	_ = json.Unmarshal(m.Extensions, &r.Extensions)
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

// Model builds a model.
func (r *TaskGroup) Model() (m *model.TaskGroup) {
	m = &model.TaskGroup{
		Name:  r.Name,
		Kind:  r.Kind,
		Addon: r.Addon,
		State: r.State,
	}
	m.ID = r.ID
	m.Data, _ = json.Marshal(StrMap(r.Data))
	m.List, _ = json.Marshal(r.Tasks)
	if r.Extensions != nil {
		m.Extensions, _ = json.Marshal(r.Extensions)
	}
	if r.Bucket != nil {
		m.BucketID = &r.Bucket.ID
	}
	for _, task := range r.Tasks {
		member := task.Model()
		m.Tasks = append(
			m.Tasks,
			*member)
	}
	return
}
