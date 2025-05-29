package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/api/filter"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/secret"
	"github.com/konveyor/tackle2-hub/trigger"
	"gorm.io/gorm"
)

// Routes
const (
	IdentitiesRoot = "/identities"
	IdentityRoot   = IdentitiesRoot + "/:" + ID
	//
	AppIdentitiesRoot = ApplicationRoot + "/identities/:" + Kind
)

// Params.
const (
	AppId = "application"
)

// IdentityHandler handles identity resource routes.
type IdentityHandler struct {
	BaseHandler
}

func (h IdentityHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("identities"))
	routeGroup.GET(IdentitiesRoot, h.List)
	routeGroup.POST(IdentitiesRoot, Transaction, h.Create)
	routeGroup.GET(IdentityRoot, h.Get)
	routeGroup.PUT(IdentityRoot, Transaction, h.Update)
	routeGroup.DELETE(IdentityRoot, h.Delete)
	//
	routeGroup.GET(AppIdentitiesRoot, h.AppList)
}

// Get godoc
// @summary Get an identity by ID.
// @description Get an identity by ID.
// @tags identities
// @produce json
// @success 200 {object} Identity
// @router /identities/{id} [get]
// @param id path int true "Identity ID"
func (h IdentityHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Identity{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Identity{}
	err := h.Decrypt(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all identities.
// @description List all identities.
// @description filters:
// @description - kind
// @description - name
// @description - application.id
// @tags dependencies
// @tags identities
// @produce json
// @success 200 {object} []Identity
// @router /identities [get]
func (h IdentityHandler) List(ctx *gin.Context) {
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "kind", Kind: qf.STRING},
			{Field: "default", Kind: qf.STRING},
			{Field: "name", Kind: qf.STRING},
			{Field: "application.id", Kind: qf.LITERAL},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Find
	var list []model.Identity
	db := h.DB(ctx)
	db = filter.Where(db)
	appFilter := filter.Resource("application")
	if !appFilter.Empty() {
		q := h.DB(ctx)
		q = q.Table("ApplicationIdentity")
		q = q.Select("IdentityID")
		appFilter = appFilter.Renamed("id", "ApplicationID")
		db = db.Where("ID IN (?)", appFilter.Where(q))
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Identity{}
	for i := range list {
		m := &list[i]
		r := Identity{}
		err := h.Decrypt(ctx, m)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		r.With(m)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AppList godoc
// @summary List application identities.
// @description List application identities.
// @tags dependencies
// @tags identities
// @produce json
// @success 200 {object} []Identity
// @router /applications/{id}/identities/{kind} [get]
func (h IdentityHandler) AppList(ctx *gin.Context) {
	id := h.pk(ctx)
	kind := ctx.Param("kind")
	var direct []model.Identity
	db := h.DB(ctx)
	db = db.Joins("JOIN ApplicationIdentity j ON j.IdentityID = Identity.ID")
	db = db.Where("j.ApplicationID", id)
	err := db.Find(&direct, "kind", kind).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	mp := map[string]Identity{}
	for i := range direct {
		m := &direct[i]
		r := Identity{}
		err := h.Decrypt(ctx, m)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		r.With(m)
		mp[r.Kind] = r
	}
	db = h.DB(ctx)
	var indirect []model.Identity
	db = db.Where("default", true)
	err = db.Find(&indirect, "kind", kind).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range indirect {
		m := &indirect[i]
		r := Identity{}
		err := h.Decrypt(ctx, m)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		r.With(m)
		mp[r.Kind] = r
	}
	resources := []Identity{}
	for _, r := range mp {
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create an identity.
// @description Create an identity.
// @tags identities
// @accept json
// @produce json
// @success 201 {object} Identity
// @router /identities [post]
// @param identity body Identity true "Identity data"
func (h IdentityHandler) Create(ctx *gin.Context) {
	r := &Identity{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if r.Default {
		defId, err := h.getDefault(ctx, r.Kind)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		if defId > 0 {
			err = &BadRequestError{
				Reason: "Kind already has default.",
			}
			_ = ctx.Error(err)
			return
		}
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	err = secret.Encrypt(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	err = db.Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	rtx := RichContext(ctx)
	tr := trigger.Identity{
		Trigger: trigger.Trigger{
			TaskManager: rtx.TaskManager,
			Client:      rtx.Client,
			DB:          h.DB(ctx),
		},
	}
	err = tr.Created(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an identity.
// @description Delete an identity.
// @tags identities
// @success 204
// @router /identities/{id} [delete]
// @param id path int true "Identity ID"
func (h IdentityHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	identity := &model.Identity{}
	result := h.DB(ctx).First(identity, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	result = h.DB(ctx).Delete(identity)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update an identity.
// @description Update an identity.
// @tags identities
// @accept json
// @success 204
// @router /identities/{id} [put]
// @param id path int true "Identity ID"
// @param identity body Identity true "Identity data"
func (h IdentityHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Identity{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	if r.Default {
		defId, err := h.getDefault(ctx, r.Kind)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		if defId > 0 && defId != id {
			err = &BadRequestError{
				Reason: "Kind already has default.",
			}
			_ = ctx.Error(err)
			return
		}
	}
	m := r.Model()
	err = secret.Encrypt(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	rtx := RichContext(ctx)
	tr := trigger.Identity{
		Trigger: trigger.Trigger{
			TaskManager: rtx.TaskManager,
			Client:      rtx.Client,
			DB:          h.DB(ctx),
		},
	}
	err = tr.Updated(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// ids return identity IDs (query) based on the filter.
func (h IdentityHandler) ids(ctx *gin.Context, f qf.Filter) (q *gorm.DB) {
	q = h.DB(ctx)
	appFilter := f.Resource("application")
	if appFilter.Empty() {
		return
	}
	q = q.Table("ApplicationIdentity")
	q = q.Select("IdentityID")
	appFilter = appFilter.Renamed("id", "ApplicationID")
	q = q.Or("ID", appFilter.Where(q))
	return
}

// getDefault returns the default by kind.
func (h IdentityHandler) getDefault(ctx *gin.Context, kind string) (id uint, err error) {
	db := h.DB(ctx)
	m := &model.Identity{}
	db = db.Model(m)
	db = db.Where("kind", kind)
	db = db.Where("default", true)
	err = db.First(m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	id = m.ID
	return
}

// Identity REST resource.
type Identity struct {
	Resource    `yaml:",inline"`
	Kind        string `json:"kind" binding:"required"`
	Default     bool   `json:"default"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Key         string `json:"key"`
	Settings    string `json:"settings"`
}

// With updates the resource with the model.
func (r *Identity) With(m *model.Identity) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Default = m.Default
	r.Name = m.Name
	r.Description = m.Description
	r.User = m.User
	r.Password = m.Password
	r.Key = m.Key
	r.Settings = m.Settings
}

// Model builds a model.
func (r *Identity) Model() (m *model.Identity) {
	m = &model.Identity{
		Kind:        r.Kind,
		Default:     r.Default,
		Name:        r.Name,
		Description: r.Description,
		User:        r.User,
		Password:    r.Password,
		Key:         r.Key,
		Settings:    r.Settings,
	}
	m.ID = r.ID

	return
}
