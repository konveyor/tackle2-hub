package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/trigger"
	"gorm.io/gorm/clause"
)

// Routes
const (
	IdentitiesRoot = "/identities"
	IdentityRoot   = IdentitiesRoot + "/:" + ID
)

// Params.
const (
	Decrypted = "decrypted"
	AppId     = "application"
)

// IdentityHandler handles identity resource routes.
type IdentityHandler struct {
	BaseHandler
}

func (h IdentityHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("identities"))
	routeGroup.GET(IdentitiesRoot, h.setDecrypted, h.List)
	routeGroup.GET(IdentitiesRoot+"/", h.setDecrypted, h.List)
	routeGroup.POST(IdentitiesRoot, h.Create)
	routeGroup.GET(IdentityRoot, h.setDecrypted, h.Get)
	routeGroup.PUT(IdentityRoot, h.Update, Transaction)
	routeGroup.DELETE(IdentityRoot, h.Delete)
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
	decrypted := ctx.GetBool(Decrypted)
	if decrypted {
		err := m.Decrypt()
		if err != nil {
			h.Status(ctx, http.StatusInternalServerError)
			return
		}
	}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all identities.
// @description List all identities.
// @tags identities
// @produce json
// @success 200 {object} []Identity
// @router /identities [get]
func (h IdentityHandler) List(ctx *gin.Context) {
	var list []model.Identity
	appId := ctx.Query(AppId)
	kind := ctx.Query(Kind)
	db := h.DB(ctx)
	if appId != "" {
		db = db.Where(
			"id IN (SELECT identityID from ApplicationIdentity WHERE applicationID = ?)",
			appId)
	}
	if kind != "" {
		db = db.Where(Kind, kind)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	decrypted := ctx.GetBool(Decrypted)
	resources := []Identity{}
	for i := range list {
		r := Identity{}
		m := &list[i]
		if decrypted {
			err := m.Decrypt()
			if err != nil {
				h.Status(ctx, http.StatusInternalServerError)
				return
			}
		}
		r.With(m)
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
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	ref := &model.Identity{}
	err = m.Encrypt(ref)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

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
	ref := &model.Identity{}
	err = h.DB(ctx).Preload(clause.Associations).First(ref, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	err = m.Encrypt(ref)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	err = db.Updates(h.fields(m)).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	rtx := WithContext(ctx)
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

// Set `decrypted` in the context.
// Results in 403 when the token does not have the required scope.
func (h *IdentityHandler) setDecrypted(ctx *gin.Context) {
	q := ctx.Query(Decrypted)
	requested, _ := strconv.ParseBool(q)
	ctx.Set(Decrypted, requested)
	if requested {
		if !h.HasScope(ctx, "identities:decrypt") {
			h.Status(ctx, http.StatusForbidden)
		} else {
			ctx.Next()
		}
	}
	return
}

// Identity REST resource.
type Identity struct {
	Resource    `yaml:",inline"`
	Kind        string `json:"kind" binding:"required"`
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
