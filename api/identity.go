package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
)

//
// Routes
const (
	IdentitiesRoot    = "/identities"
	IdentityRoot      = IdentitiesRoot + "/:" + ID
	AppIdentitiesRoot = ApplicationRoot + IdentitiesRoot
)

//
// IdentityHandler handles identity resource routes.
type IdentityHandler struct {
	BaseHandler
}

func (h IdentityHandler) AddRoutes(e *gin.Engine) {
	e.GET(IdentitiesRoot, h.List)
	e.GET(IdentitiesRoot+"/", h.List)
	e.POST(IdentitiesRoot, h.Create)
	e.GET(IdentityRoot, h.Get)
	e.PUT(IdentityRoot, h.Update)
	e.DELETE(IdentityRoot, h.Delete)
	e.GET(AppIdentitiesRoot, h.ListByApplication)
	e.GET(AppIdentitiesRoot+"/", h.ListByApplication)
}

// Get godoc
// @summary Get an identity by ID.
// @description Get an identity by ID.
// @tags get
// @produce json
// @success 200 {object} Identity
// @router /identities/{id} [get]
// @param id path string true "Identity ID"
func (h IdentityHandler) Get(ctx *gin.Context) {
	m := &model.Identity{}
	id := ctx.Param(ID)
	result := h.DB.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Identity{}
	r.With(m)

	ctx.JSON(http.StatusOK, r)
}

// List godoc
// @summary List all identities.
// @description List all identities.
// @tags get
// @produce json
// @success 200 {object} []Identity
// @router /identities [get]
func (h IdentityHandler) List(ctx *gin.Context) {
	var list []model.Identity
	result := h.DB.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Identity{}
	for i := range list {
		r := Identity{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create an identity.
// @description Create an identity.
// @tags create
// @accept json
// @produce json
// @success 201 {object} Identity
// @router /identities [post]
// @param identity body Identity true "Identity data"
func (h IdentityHandler) Create(ctx *gin.Context) {
	r := &Identity{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	result := h.DB.Create(m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	r.With(m)

	ctx.JSON(http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an identity.
// @description Delete an identity.
// @tags delete
// @success 204
// @router /identities/{id} [delete]
// @param id path string true "Identity ID"
func (h IdentityHandler) Delete(ctx *gin.Context) {
	id := ctx.Param(ID)
	identity := &model.Identity{}
	result := h.DB.First(identity, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}
	result = h.DB.Delete(identity, id)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Update godoc
// @summary Update an identity.
// @description Update an identity.
// @tags update
// @accept json
// @success 204
// @router /identities/{id} [put]
// @param id path string true "Identity ID"
// @param identity body Identity true "Identity data"
func (h IdentityHandler) Update(ctx *gin.Context) {
	id := ctx.Param(ID)
	r := &Identity{}
	err := ctx.BindJSON(r)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}
	m := r.Model()
	db := h.DB.Model(m)
	db = db.Where("id", id)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListByApplication  godoc
// @summary List identities for an application.
// @description List identities for an application.
// @tags get
// @produce json
// @success 200 {object} []Identity
// @router /application-inventory/application/{id}/identities [get]
// @param id path int true "Application ID"
func (h IdentityHandler) ListByApplication(ctx *gin.Context) {
	m := &model.Application{}
	id := ctx.Param(ID)
	db := h.preLoad(h.DB, "Identities")
	result := db.First(m, id)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	resources := []Identity{}
	for i := range m.Identities {
		id := Identity{}
		id.With(&m.Identities[i])
		resources = append(
			resources,
			id)
	}

	ctx.JSON(http.StatusOK, resources)
}

//
// Identity REST resource.
type Identity struct {
	Resource
	Kind        string `json:"kind" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Key         string `json:"key"`
	Settings    string `json:"settings"`
	Encrypted   string `json:"encrypted"`
}

//
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
	r.Encrypted = m.Encrypted
}

//
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
		Encrypted:   r.Encrypted,
	}
	m.ID = r.ID

	return
}
