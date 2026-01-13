package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TaskGroupHandler handles task group routes.
type TaskGroupHandler struct {
	BucketOwner
}

// AddRoutes adds routes.
func (h TaskGroupHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tasks"), Transaction)
	routeGroup.GET(api.TaskGroupsRoute, h.List)
	routeGroup.GET(api.TaskGroupsRoute+"/", h.List)
	routeGroup.POST(api.TaskGroupsRoute, h.Create)
	routeGroup.PUT(api.TaskGroupRoute, h.Update)
	routeGroup.PATCH(api.TaskGroupRoute, Transaction, h.Update)
	routeGroup.GET(api.TaskGroupRoute, h.Get)
	routeGroup.PUT(api.TaskGroupSubmitRoute, Transaction, h.Submit)
	routeGroup.DELETE(api.TaskGroupRoute, h.Delete)
	// Bucket
	routeGroup = e.Group("/")
	routeGroup.Use(Required("tasks.bucket"))
	routeGroup.GET(api.TaskGroupBucketRoute, h.BucketGet)
	routeGroup.GET(api.TaskGroupBucketContentRoute, h.BucketGet)
	routeGroup.POST(api.TaskGroupBucketContentRoute, h.BucketPut)
	routeGroup.PUT(api.TaskGroupBucketContentRoute, h.BucketPut)
	routeGroup.DELETE(api.TaskGroupBucketContentRoute, h.BucketDelete)
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
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
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
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
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
	rtx := RichContext(ctx)
	m := &model.TaskGroup{}
	r.Patch(m)
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx)
	db = db.Omit(clause.Associations)
	switch r.State {
	case "":
		m.State = task.Created
		fallthrough
	case task.Created:
		err = db.Create(&m).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	case task.Ready:
		taskGroup := task.NewTaskGroup(m)
		err = taskGroup.Submit(h.DB(ctx), rtx.TaskManager)
		if err != nil {
			_ = ctx.Error(err)
			return
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
		r.State = task.Ready
	}
	err = h.findRefs(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Patch(m)
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	switch m.State {
	case "", task.Created:
		db := h.DB(ctx)
		db = db.Omit(
			clause.Associations,
			"BucketID",
			"Bucket")
		err = db.Save(m).Error
		if err != nil {
			_ = ctx.Error(err)
			return
		}
	case task.Ready:
		rtx := RichContext(ctx)
		taskGroup := task.NewTaskGroup(m)
		err = taskGroup.Submit(h.DB(ctx), rtx.TaskManager)
		if err != nil {
			_ = ctx.Error(err)
			return
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
	rtx := RichContext(ctx)
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
// @tags taskgroups
// @produce json
// @success 204
// @router /taskgroups/{id}/bucket/{wildcard} [post]
// @param id path int true "TaskGroup ID"
// @param wildcard path string true "Content path"
func (h TaskGroupHandler) BucketPut(ctx *gin.Context) {
	m := &model.TaskGroup{}
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
// @tags taskgroups
// @produce json
// @success 204
// @router /taskgroups/{id}/bucket/{wildcard} [delete]
// @param id path int true "Task ID"
// @param wildcard path string true "Content path"
func (h TaskGroupHandler) BucketDelete(ctx *gin.Context) {
	m := &model.TaskGroup{}
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
		data := model.Data{Any: r.Data}
		other := model.Data{Any: kind.Data()}
		merged := data.Merge(other)
		if !merged {
			r.Data = other.Any
		}
	}
	return
}

// TaskGroup REST resource.
type TaskGroup = resource.TaskGroup
