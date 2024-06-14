package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
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
	routeGroup.PATCH(TaskGroupRoot, Transaction, h.Update)
	routeGroup.GET(TaskGroupRoot, h.Get)
	routeGroup.PUT(TaskGroupSubmitRoot, Transaction, h.Submit)
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
	err = h.findRefs(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Omit(clause.Associations)
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	switch r.State {
	case "":
		m.State = tasking.Created
		fallthrough
	case tasking.Created:
		result := db.Create(&m)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
	case tasking.Ready:
		err := h.Propagate(m)
		if err != nil {
			return
		}
		result := db.Create(&m)
		if result.Error != nil {
			_ = ctx.Error(result.Error)
			return
		}
		rtx := WithContext(ctx)
		for i := range m.Tasks {
			task := &tasking.Task{}
			task.With(&m.Tasks[i])
			err = rtx.TaskManager.Create(h.DB(ctx), task)
			if err != nil {
				_ = ctx.Error(err)
				return
			}
		}
	default:
		_ = ctx.Error(
			&BadRequestError{
				Reason: "state must be ('''|Created|Ready)",
			})
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
	m := &model.TaskGroup{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r := &TaskGroup{}
	if ctx.Request.Method == http.MethodPatch &&
		ctx.Request.ContentLength > 0 {
		r.With(m)
	}
	err = h.Bind(ctx, r)
	if err != nil {
		return
	}
	if _, found := ctx.Get(Submit); found {
		r.State = tasking.Ready
	}
	err = h.findRefs(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	db = db.Omit(
		clause.Associations,
		"BucketID",
		"Bucket")
	m = r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	switch m.State {
	case "", tasking.Created:
		err = db.Save(m).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	case tasking.Ready:
		for i := range m.Tasks {
			task := &m.Tasks[i]
			if task.ID > 0 {
				_ = ctx.Error(
					&BadRequestError{
						Reason: "already submitted.",
					})
				return
			}
		}
		err := h.Propagate(m)
		if err != nil {
			return
		}
		err = db.Save(m).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		rtx := WithContext(ctx)
		for i := range m.Tasks {
			task := &tasking.Task{}
			task.With(&m.Tasks[i])
			task.CreateUser = h.CurrentUser(ctx)
			err = rtx.TaskManager.Create(h.DB(ctx), task)
			if err != nil {
				_ = ctx.Error(err)
				return
			}
		}
	default:
		_ = ctx.Error(
			&BadRequestError{
				Reason: "state must be (Created|Ready)",
			})
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
	db := h.DB(ctx)
	db = db.Omit(clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	rtx := WithContext(ctx)
	for i := range m.Tasks {
		task := &m.Tasks[i]
		err = rtx.TaskManager.Delete(h.DB(ctx), task.ID)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	}
	err = db.Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Submit godoc
// @summary Submit a task group.
// @description Patch and submit a task group.
// @tags taskgroups
// @accept json
// @success 204
// @router /taskgroups/{id}/submit [put]
// @param id path int true "TaskGroup ID"
// @param taskgroup body TaskGroup false "TaskGroup data (optional)"
func (h TaskGroupHandler) Submit(ctx *gin.Context) {
	ctx.Set(Submit, true)
	ctx.Request.Method = http.MethodPatch
	h.Update(ctx)
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

// findRefs find referenced resources.
// - addon
// - extensions
// - kind
// - priority
// The priority is defaulted to the kind as needed.
func (h *TaskGroupHandler) findRefs(ctx *gin.Context, r *TaskGroup) (err error) {
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
				Name:      r.Kind,
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
		mA, castA := h.AsMap(kind.Spec.Data)
		mB, castB := r.Data.(map[string]any)
		if castA && castB {
			r.Data = h.Merge(mA, mB)
		} else {
			r.Data = mA
		}
	}
	return
}

// Propagate group data into the task.
func (h *TaskGroupHandler) Propagate(m *model.TaskGroup) (err error) {
	for i := range m.Tasks {
		task := &m.Tasks[i]
		task.Kind = m.Kind
		task.Addon = m.Addon
		task.Extensions = m.Extensions
		task.Priority = m.Priority
		task.Policy = m.Policy
		task.State = m.State
		task.SetBucket(m.BucketID)
		if m.Data.Any != nil {
			mA, castA := m.Data.Any.(map[string]any)
			mB, castB := task.Data.Any.(map[string]any)
			if castA && castB {
				task.Data.Any = h.Merge(mA, mB)
			} else {
				task.Data.Any = m.Data
			}
		}
	}

	return
}

// TaskGroup REST resource.
type TaskGroup struct {
	Resource   `yaml:",inline"`
	Name       string     `json:"name"`
	Kind       string     `json:"kind,omitempty" yaml:",omitempty"`
	Addon      string     `json:"addon,omitempty" yaml:",omitempty"`
	Extensions []string   `json:"extensions,omitempty" yaml:",omitempty"`
	State      string     `json:"state"`
	Priority   int        `json:"priority,omitempty" yaml:",omitempty"`
	Policy     TaskPolicy `json:"policy,omitempty" yaml:",omitempty"`
	Data       any        `json:"data" swaggertype:"object" binding:"required"`
	Bucket     *Ref       `json:"bucket,omitempty"`
	Tasks      []Task     `json:"tasks"`
}

// With updates the resource with the model.
func (r *TaskGroup) With(m *model.TaskGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Kind = m.Kind
	r.Addon = m.Addon
	r.Extensions = m.Extensions
	r.State = m.State
	r.Priority = m.Priority
	r.Policy = TaskPolicy(m.Policy)
	r.Data = m.Data.Any
	r.Bucket = r.refPtr(m.BucketID, m.Bucket)
	r.Tasks = []Task{}
	switch m.State {
	case "", tasking.Created:
		for _, task := range m.List {
			member := Task{}
			member.With(&task)
			r.Tasks = append(
				r.Tasks,
				member)
		}
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
		Name:       r.Name,
		Kind:       r.Kind,
		Addon:      r.Addon,
		Extensions: r.Extensions,
		State:      r.State,
		Priority:   r.Priority,
		Policy:     model.TaskPolicy(r.Policy),
	}
	m.ID = r.ID
	m.Data.Any = r.Data
	for _, task := range r.Tasks {
		m.List = append(m.List, *task.Model())
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
