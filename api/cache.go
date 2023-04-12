package api

import (
	"github.com/gin-gonic/gin"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/nas"
	"net/http"
	"os"
	"os/exec"
	pathlib "path"
	"strings"
)

//
// Routes
const (
	CacheRoot    = "/cache"
	CacheDirRoot = CacheRoot + "/*" + Wildcard
)

//
// CacheHandler handles cache routes.
type CacheHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h CacheHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("cache"))
	routeGroup.GET(CacheRoot, h.Get)
	routeGroup.GET(CacheDirRoot, h.Get)
	routeGroup.DELETE(CacheDirRoot, h.Delete)
}

// Get godoc
// @summary Get the cache.
// @description Get the cache.
// @tags cache
// @produce json
// @success 200 {object} api.Cache
// @router /caches/{id} [get]
// @param name path string true "Cache DIR"
func (h CacheHandler) Get(ctx *gin.Context) {
	dir := ctx.Param(Wildcard)
	r, err := h.cache(dir)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.Status(http.StatusNotFound)
		} else {
			_ = ctx.Error(err)
		}
		return
	}

	h.Render(ctx, http.StatusOK, r)
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
		ctx.Status(http.StatusForbidden)
		return
	}
	path := pathlib.Join(
		Settings.Cache.Path,
		dir)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.Status(http.StatusNoContent)
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

	ctx.Status(http.StatusNoContent)
}

//
// cache builds the resource.
func (h *CacheHandler) cache(dir string) (cache *Cache, err error) {
	cache = &Cache{}
	cache.Path = pathlib.Join(
		Settings.Cache.Path,
		dir)
	_, err = os.Stat(Settings.Cache.Path)
	if err != nil {
		return
	}
	cmd := exec.Command("/usr/bin/df", "-h", Settings.Cache.Path)
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

//
// Cache REST resource.
type Cache struct {
	Path     string `json:"path"`
	Capacity string `json:"capacity"`
	Used     string `json:"used"`
	Exists   bool   `json:"exists"`
}
