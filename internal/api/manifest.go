package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	qf "github.com/konveyor/tackle2-hub/internal/api/filter"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
)

const (
	Injected = api.Injected
)

var SecretRefPattern = regexp.MustCompile("\\$\\(([^)]+)\\)")

// ManifestHandler handles application Manifest resource routes.
type ManifestHandler struct {
	BaseHandler
}

func (h ManifestHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("manifests"))
	routeGroup.GET(api.ManifestRoute, h.Get)
	routeGroup.GET(api.ManifestsRoute, h.List)
	routeGroup.GET(api.ManifestsRoute+"/", h.List)
	routeGroup.POST(api.ManifestsRoute, h.Create)
	routeGroup.PUT(api.ManifestRoute, h.Update)
	routeGroup.DELETE(api.ManifestRoute, h.Delete)
	// application
	routeGroup = e.Group("/")
	routeGroup.Use(Required("applications.manifests"))
	routeGroup.GET(api.AppManifestRoute, h.AppGet)
	routeGroup.POST(api.AppManifestsRoute, h.AppCreate)
}

// Get godoc
// @summary Get a Manifest by ID.
// @description Get a Manifest by ID.
// @tags manifests
// @produce json
// @success 200 {object} Manifest
// @router /manifests/{id} [get]
// @param id path int true "Manifest ID"
// @Param injected query bool false "Inject secret references"
// @Param decrypted query bool false "Decrypt secret references"
func (h ManifestHandler) Get(ctx *gin.Context) {
	r := Manifest{}
	id := h.pk(ctx)
	m := &model.Manifest{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Decrypt(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.inject(ctx, &r)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all manifests.
// @description List all manifests.
// @description filters:
// @description   - application.id
// @tags manifests
// @produce json
// @success 200 {object} []Manifest
// @router /manifests [get]
// @Param injected query bool false "Inject secret references"
// @Param decrypted query bool false "Decrypt secret references"
func (h ManifestHandler) List(ctx *gin.Context) {
	resources := []Manifest{}
	// Filter
	filter, err := qf.New(ctx,
		[]qf.Assert{
			{Field: "application.id", Kind: qf.LITERAL},
		})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	filter = filter.Renamed("application.id", "applicationid")
	// Fetch.
	var list []model.Manifest
	db := h.DB(ctx)
	db = filter.Where(db)
	err = db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	for i := range list {
		m := &list[i]
		err := h.Decrypt(ctx, m)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		r := Manifest{}
		r.With(m)
		h.inject(ctx, &r)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a manifest.
// @description Create a manifest.
// @tags manifests
// @accept json
// @produce json
// @success 201 {object} Manifest
// @router /manifests [post]
// @param manifest body Manifest true "Manifest data"
func (h ManifestHandler) Create(ctx *gin.Context) {
	r := &Manifest{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	err = secret.Encrypt(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m.CreateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a manifest.
// @description Delete a manifest.
// @tags manifests
// @success 204
// @router /manifests/{id} [delete]
// @param id path int true "Manifest ID"
func (h ManifestHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Manifest{}
	db := h.DB(ctx)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = db.Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a manifest.
// @description Update a manifest.
// @tags manifests
// @accept json
// @success 204
// @router /manifests/{id} [put]
// @param id path int true "Manifest ID"
// @param manifest body Manifest true "Manifest data"
func (h ManifestHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Manifest{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	err = secret.Encrypt(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db := h.DB(ctx)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// AppGet godoc
// @summary Get the latest application Manifest.
// @description Get the latest application Manifest.
// @tags manifests
// @produce json
// @success 200 {object} Manifest
// @router /applications/{id}/manifest [get]
// @param id path int true "Application ID"
// @Param injected query bool false "Inject secret references"
// @Param decrypted query bool false "Decrypt secret references"
func (h *ManifestHandler) AppGet(ctx *gin.Context) {
	appId := h.pk(ctx)
	r := Manifest{}
	m := &model.Manifest{}
	db := h.DB(ctx)
	err := db.Last(m, "ApplicationID", appId).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.Decrypt(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.inject(ctx, &r)

	h.Respond(ctx, http.StatusOK, r)
}

// AppCreate godoc
// @summary Create a manifest.
// @description Create a manifest.
// @tags manifests
// @accept json
// @produce json
// @success 201 {object} Manifest
// @router /applications/{id}/manifests [post]
// @param id path int true "Application ID"
// @param manifest body Manifest true "Manifest data"
func (h ManifestHandler) AppCreate(ctx *gin.Context) {
	appId := h.pk(ctx)
	r := &Manifest{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	err = secret.Encrypt(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m.ApplicationID = appId
	m.CreateUser = h.CurrentUser(ctx)
	db := h.DB(ctx)
	err = db.Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// inject replaces secret refs with values.
func (h *ManifestHandler) inject(ctx *gin.Context, r *Manifest) {
	q := ctx.Query(Injected)
	injected, _ := strconv.ParseBool(q)
	if !injected {
		return
	}
	var inject func(m resource.Map)
	inject = func(m resource.Map) {
		for k, v := range m {
			switch object := v.(type) {
			case string:
				matched := SecretRefPattern.FindAllStringSubmatch(object, -1)
				for _, match := range matched {
					if len(match) != 2 {
						break
					}
					ref := match[1]
					sv, found := r.Secret[ref]
					if !found {
						break
					}
					s, cast := sv.(string)
					if !cast {
						break
					}
					object = strings.Replace(
						object,
						match[0],
						s,
						-1)
				}
				m[k] = object
			case resource.Map:
				inject(object)
			case map[string]any:
				inject(object)
			}
		}
	}
	inject(r.Content)
}

// Manifest REST resource.
type Manifest = resource.Manifest
