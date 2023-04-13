package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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
	routeGroup.Use(Required("settings"))
	routeGroup.GET(SettingsRoot, h.List)
	routeGroup.GET(SettingsRoot+"/", h.List)
	routeGroup.GET(SettingRoot, h.Get)
	routeGroup.POST(SettingsRoot, h.Create)
	routeGroup.POST(SettingRoot, h.CreateByKey)
	routeGroup.PUT(SettingRoot, h.Update)
	routeGroup.DELETE(SettingRoot, h.Delete)
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

	h.Render(ctx, http.StatusOK, r.Value)
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
	result := h.Paginated(ctx).Find(&list)
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

	h.Render(ctx, http.StatusOK, resources)
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
		h.Render(ctx,
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

	h.Render(ctx, http.StatusCreated, setting)
}

// CreateByKey godoc
// @summary Create a setting.
// @description Create a setting.
// @tags settings
// @accept json
// @success 201
// @router /settings/{key} [post]
// @param setting body api.Setting true "Setting value"
func (h SettingHandler) CreateByKey(ctx *gin.Context) {
	key := ctx.Param(Key)
	if strings.HasPrefix(key, ".") {
		h.Render(ctx,
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

	ctx.Status(http.StatusCreated)
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
		h.Render(ctx,
			http.StatusForbidden,
			gin.H{
				"error": fmt.Sprintf("%s is read-only.", key),
			})

		return
	}

	updates := Setting{}
	updates.Key = key
	err := h.Bind(ctx, &updates.Value)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	m := updates.Model()
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Where("key", key)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}

	ctx.Status(http.StatusNoContent)
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
		h.Render(ctx,
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
