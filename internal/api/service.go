package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// serviceRoutes name to route map.
var serviceRoutes = map[string]string{
	"llm-proxy": os.Getenv("LLM_PROXY_URL"),
	"kai":       os.Getenv("KAI_URL"),
}

// ServiceHandler handles service routes.
type ServiceHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h ServiceHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Authenticate())
	routeGroup.GET(api.ServicesRoute, h.List)
	routeGroup.Any(api.ServiceRoute, h.Forward)
	routeGroup.Any(api.ServiceNestedRoute, h.Forward)
}

// List godoc
// @summary List named service routes.
// @description List named service routes.
// @tags services
// @produce json
// @success 200 {object} api.Service
// @router /services [get]
func (h ServiceHandler) List(ctx *gin.Context) {
	var r []Service
	for name, route := range serviceRoutes {
		service := Service{Name: name, Route: route}
		r = append(r, service)
	}

	h.Respond(ctx, http.StatusOK, r)
}

// Forward provides RBAC and forwards request to the service.
func (h ServiceHandler) Forward(ctx *gin.Context) {
	path := ctx.Param(Wildcard)
	name := ctx.Param(Name)
	route, found := serviceRoutes[name]
	if !found {
		err := &NotFound{Resource: name}
		_ = ctx.Error(err)
		return
	}
	if route == "" {
		err := fmt.Errorf("route for: '%s' not defined", name)
		_ = ctx.Error(err)
		return
	}
	u, err := url.Parse(route)
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
		_ = ctx.Error(err)
		return
	}
	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = path
			Log.Info(
				"Routing (service)",
				"path",
				ctx.Request.URL.Path,
				"route",
				req.URL.String())
		},
		FlushInterval: -1, // Flush immediately for SSE/streaming (MCP, LLM responses)
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			if req.Context().Err() != nil {
				return
			}
			Log.Error(err, "Proxy error", "path", req.URL.Path)
		},
	}

	// httputil.ReverseProxy panics with http.ErrAbortHandler when an SSE/streaming
	// client disconnects mid-response. Recover gracefully instead of crashing the
	// request pipeline.
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				Log.V(1).Info("Client disconnected during proxy streaming",
					"path", ctx.Request.URL.Path)
				return
			}
			panic(r)
		}
	}()

	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}

// Service REST resource.
type Service = resource.Service
