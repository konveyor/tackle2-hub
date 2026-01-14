package api

import (
	"net/http"
	"os"
	"os/exec"
	pathlib "path"
	"strings"

	"github.com/gin-gonic/gin"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/nas"
)

// CacheHandler handles cache routes.
type CacheHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h CacheHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("cache"))
	routeGroup.GET(api.CacheRoute, h.Get)
	routeGroup.GET(api.CacheDirRoute, h.Get)
	routeGroup.DELETE(api.CacheDirRoute, h.Delete)
}

// Get godoc
// @summary Get the cache.
// @description Get the cache.
// @tags cache
// @produce json
// @success 200 {object} api.Cache
// @router /caches/{wildcard} [get]
// @param wildcard path string true "Cache DIR"
func (h CacheHandler) Get(ctx *gin.Context) {
	dir := ctx.Param(Wildcard)
	r, err := h.cache(dir)
	if err != nil {
		if os.IsNotExist(err) {
			h.Status(ctx, http.StatusNotFound)
		} else {
			_ = ctx.Error(err)
		}
		return
	}

	h.Respond(ctx, http.StatusOK, r)
}

// Delete godoc
// @summary Delete a directory within the cache.
// @description Delete a directory within the cache.
// @tags cache
// @produce json
// @success 204
// @router /cache [delete]
func (h CacheHandler) Delete(ctx *gin.Context) {
	dir := ctx.Param(Wildcard)
	if dir == "" {
		h.Status(ctx, http.StatusForbidden)
		return
	}
	path := pathlib.Join(
		Settings.Addon.CacheDir,
		dir)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			h.Status(ctx, http.StatusNoContent)
		} else {
			_ = ctx.Error(err)
		}
		return
	}
	err = nas.RmDir(path)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// cache builds the resource.
func (h *CacheHandler) cache(dir string) (cache *Cache, err error) {
	cache = &Cache{}
	cache.Path = pathlib.Join(
		Settings.Addon.CacheDir,
		dir)
	_, err = os.Stat(Settings.Addon.CacheDir)
	if err != nil {
		return
	}
	cmd := exec.Command("/usr/bin/df", "-h", Settings.Addon.CacheDir)
	stdout, err := cmd.Output()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	output := string(stdout)
	lines := strings.Split(output, "\n")

	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return
	}
	cache.Capacity = fields[1]
	cache.Used = fields[2]
	cache.Exists = true
	if dir == "" {
		return
	}
	_, err = os.Stat(cache.Path)
	if err != nil {
		cache.Used = "0"
		cache.Exists = false
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	cmd = exec.Command("/usr/bin/du", "-h", "-d0", cache.Path)
	stdout, err = cmd.Output()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	output = string(stdout)
	lines = strings.Split(output, "\n")
	fields = strings.Fields(lines[0])
	cache.Used = fields[0]
	return
}

// Cache REST resource.
type Cache = resource.Cache
