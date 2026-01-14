package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// SettingHandler handles setting routes.
type SettingHandler struct {
	BaseHandler
}

// AddRoutes add routes.
func (h SettingHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("settings"))
	routeGroup.GET(api.SettingsRoute, h.List)
	routeGroup.GET(api.SettingsRoute+"/", h.List)
	routeGroup.GET(api.SettingRoute, h.Get)
	routeGroup.POST(api.SettingsRoute, h.Create)
	routeGroup.POST(api.SettingRoute, h.CreateByKey)
	routeGroup.PUT(api.SettingRoute, h.Update)
	routeGroup.DELETE(api.SettingRoute, h.Delete)
}

// Get godoc
// @summary Get a setting by its key.
// @description Get a setting by its key.
// @tags settings
// @produce json
// @success 200 {object} api.Setting
// @router /settings/{key} [get]
// @param key path string true "Key"
func (h SettingHandler) Get(ctx *gin.Context) {
	setting := &model.Setting{}
	key := ctx.Param(Key)
	result := h.DB(ctx).Where(&model.Setting{Key: key}).First(setting)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Setting{}
	r.With(setting)

	h.Respond(ctx, http.StatusOK, r.Value)
}

// List godoc
// @summary List all settings.
// @description List all settings.
// @tags settings
// @produce json
// @success 200 array api.Setting
// @router /settings [get]
func (h SettingHandler) List(ctx *gin.Context) {
	var list []model.Setting
	result := h.DB(ctx).Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Setting{}
	for i := range list {
		r := Setting{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a setting.
// @description Create a setting.
// @tags settings
// @accept json
// @produce json
// @success 201 {object} api.Setting
// @router /settings [post]
// @param setting body api.Setting true "Setting data"
func (h SettingHandler) Create(ctx *gin.Context) {
	setting := Setting{}
	err := h.Bind(ctx, &setting)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	if strings.HasPrefix(setting.Key, ".") {
		h.Respond(ctx,
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", setting.Key),
			})

		return
	}

	m := setting.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(&m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	setting.With(m)

	h.Respond(ctx, http.StatusCreated, setting)
}

// CreateByKey godoc
// @summary Create a setting.
// @description Create a setting.
// @tags settings
// @accept json
// @success 201
// @router /settings/{key} [post]
// @param key path string true "Key"
// @param setting body api.Setting true "Setting value"
func (h SettingHandler) CreateByKey(ctx *gin.Context) {
	key := ctx.Param(Key)
	if strings.HasPrefix(key, ".") {
		h.Respond(ctx,
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", key),
			})

		return
	}

	setting := Setting{}
	setting.Key = key
	err := h.Bind(ctx, &setting.Value)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := setting.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(&m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusCreated)
}

// Update godoc
// @summary Update a setting.
// @description Update a setting.
// @tags settings
// @accept json
// @produce json
// @success 204
// @router /settings/{key} [put]
// @param key path string true "Key"
func (h SettingHandler) Update(ctx *gin.Context) {
	key := ctx.Param(Key)
	if strings.HasPrefix(key, ".") {
		h.Respond(ctx,
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", key),
			})

		return
	}

	m := &model.Setting{}
	result := h.DB(ctx).First(m, "key = ?", key)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err := h.Bind(ctx, &m.Value)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	result = h.DB(ctx).Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}

	h.Status(ctx, http.StatusNoContent)
}

// Delete godoc
// @summary Delete a setting.
// @description Delete a setting.
// @tags settings
// @success 204
// @router /settings/{key} [delete]
// @param key path string true "Key"
func (h SettingHandler) Delete(ctx *gin.Context) {
	key := ctx.Param(Key)
	if strings.HasPrefix(key, ".") {
		h.Respond(ctx,
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", key),
			})

		return
	}

	result := h.DB(ctx).Delete(&model.Setting{}, Key, key)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Setting REST Resource
type Setting = resource.Setting
