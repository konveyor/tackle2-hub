package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/auth"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

//
// Routes
const (
	PathfinderRoot   = "/pathfinder"
	AssessmentsRoot  = "assessments"
	AssessmentsRootX = AssessmentsRoot + "/*" + Wildcard
)

//
// PathfinderHandler handles assessment routes.
type PathfinderHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h PathfinderHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group(PathfinderRoot)
	routeGroup.Use(auth.AuthorizationRequired(h.AuthProvider, AssessmentsRoot))
	routeGroup.Any(AssessmentsRoot, h.ReverseProxy)
	routeGroup.Any(AssessmentsRootX, h.ReverseProxy)
}

// Get godoc
// @summary ReverseProxy - forward to pathfinder.
// @description ReverseProxy forwards API calls to pathfinder API.
func (h PathfinderHandler) ReverseProxy(ctx *gin.Context) {
	pathfinder := os.Getenv("PATHFINDER_URL")
	target, _ := url.Parse(pathfinder)
	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
		},
	}

	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}
