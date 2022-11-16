package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"net/http"
	"strings"
)

//
// Routes
const (
	SettingsRoot = "/settings"
	SettingRoot  = SettingsRoot + "/:" + Key
)

//
// SettingHandler handles setting routes.
type SettingHandler struct {
	BaseHandler
}

//
// AddRoutes add routes.
func (h SettingHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(auth.Required("settings"))
	routeGroup.GET(SettingsRoot, h.List)
	routeGroup.GET(SettingsRoot+"/", h.List)
	routeGroup.GET(SettingRoot, h.Get)
	routeGroup.POST(SettingsRoot, h.Create)
	routeGroup.PUT(SettingRoot, h.Update)
	routeGroup.DELETE(SettingRoot, h.Delete)
}

// Get godoc
// @summary Get a setting by its key.
// @description Get a setting by its key.
// @tags get, setting
// @produce json
// @success 200 {object} interface{}
// @router /settings/{key} [get]
// @param key path string true "Key"
func (h SettingHandler) Get(ctx *gin.Context) {
	setting := &model.Setting{}
	key := ctx.Param(Key)
	result := h.DB.Where(&model.Setting{Key: key}).First(setting)
	if result.Error != nil {
		h.getFailed(ctx, result.Error)
		return
	}
	r := Setting{}
	r.With(setting)

	ctx.JSON(http.StatusOK, r.Value)
}

// List godoc
// @summary List all settings.
// @description List all settings.
// @tags list, setting
// @produce json
// @success 200 array api.Setting
// @router /settings [get]
func (h SettingHandler) List(ctx *gin.Context) {
	var list []model.Setting
	result := h.DB.Find(&list)
	if result.Error != nil {
		h.listFailed(ctx, result.Error)
		return
	}
	resources := []Setting{}
	for i := range list {
		r := Setting{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	ctx.JSON(http.StatusOK, resources)
}

// Create godoc
// @summary Create a setting.
// @description Create a setting.
// @tags create, setting
// @accept json
// @produce json
// @success 201 {object} api.Setting
// @router /settings [post]
// @param setting body api.Setting true "Setting data"
func (h SettingHandler) Create(ctx *gin.Context) {
	setting := Setting{}
	err := ctx.BindJSON(&setting)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}

	if strings.HasPrefix(setting.Key, ".") {
		ctx.JSON(
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", setting.Key),
			})

		return
	}

	m := setting.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB.Create(&m)
	if result.Error != nil {
		h.createFailed(ctx, result.Error)
		return
	}
	setting.With(m)

	ctx.JSON(http.StatusCreated, setting)
}

// Update godoc
// @summary Update a setting.
// @description Update a setting.
// @tags update, setting
// @accept json
// @produce json
// @success 204
// @router /settings/{key} [put]
// @param key path string true "Key"
func (h SettingHandler) Update(ctx *gin.Context) {
	key := ctx.Param(Key)
	if strings.HasPrefix(key, ".") {
		ctx.JSON(
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", key),
			})

		return
	}

	updates := Setting{}
	updates.Key = key
	err := ctx.BindJSON(&updates.Value)
	if err != nil {
		h.bindFailed(ctx, err)
		return
	}

	m := updates.Model()
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB.Model(m)
	db = db.Where("key", key)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		h.updateFailed(ctx, result.Error)
	}

	ctx.Status(http.StatusNoContent)
}

// Delete godoc
// @summary Delete a setting.
// @description Delete a setting.
// @tags delete, setting
// @success 204
// @router /settings/{key} [delete]
// @param key path string true "Key"
func (h SettingHandler) Delete(ctx *gin.Context) {
	key := ctx.Param(Key)
	if strings.HasPrefix(key, ".") {
		ctx.JSON(
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", key),
			})

		return
	}

	result := h.DB.Delete(&model.Setting{}, Key, key)
	if result.Error != nil {
		h.deleteFailed(ctx, result.Error)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// Setting REST Resource
type Setting struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (r *Setting) With(m *model.Setting) {
	r.Key = m.Key
	_ = json.Unmarshal(m.Value, &r.Value)

}

func (r *Setting) Model() (m *model.Setting) {
	m = &model.Setting{Key: r.Key}
	m.Value, _ = json.Marshal(r.Value)
	return
}
