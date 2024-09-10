package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

// Routes
const (
	ServicesRoot = "/services"
	ServiceRoot  = ServicesRoot + "/:name/*" + Wildcard
)

// serviceRoutes name to route map.
var serviceRoutes = map[string]string{
	"kai": os.Getenv("KAI_URL"),
}

// ServiceHandler handles service routes.
type ServiceHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h ServiceHandler) AddRoutes(e *gin.Engine) {
	e.GET(ServicesRoot, h.List)
	e.Any(ServiceRoot, h.Required, h.Forward)
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

// Required enforces RBAC.
func (h ServiceHandler) Required(ctx *gin.Context) {
	Required(ctx.Param(Name))
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
	}

	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}

// Service REST resource.
type Service struct {
	Name  string `json:"name"`
	Route string `json:"route"`
}
