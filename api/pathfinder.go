package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/auth"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
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

//
// Pathfinder client.
type Pathfinder struct {
}

//
// DeleteAssessment deletes associated assessments by application Ids.
func (r *Pathfinder) DeleteAssessment(ids []uint, ctx *gin.Context) (err error) {
	client := r.client()
	body := map[string][]uint{"applicationIds": ids}
	b, _ := json.Marshal(body)
	header := http.Header{
		Authorization: ctx.Request.Header[Authorization],
		ContentLength: []string{strconv.Itoa(len(b))},
		ContentType:   []string{"application/json"},
	}
	request := r.request(
		http.MethodDelete,
		"bulkDelete",
		header)
	reader := bytes.NewReader(b)
	request.Body = ioutil.NopCloser(reader)
	result, err := client.Do(request)
	if err != nil {
		return
	}
	status := result.StatusCode
	switch status {
	case http.StatusNoContent:
		Log.Info(
			"Assessment(s) deleted for applications.",
			"Ids",
			ids)
	default:
		err = liberr.New(http.StatusText(status))
	}

	return
}

//
// Build the client.
func (r *Pathfinder) client() (client *http.Client) {
	client = &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	return
}

//
// Build the request
func (r *Pathfinder) request(method, endpoint string, header http.Header) (request *http.Request) {
	u, _ := url.Parse(os.Getenv("PATHFINDER_URL"))
	u.Path = path.Join(
		u.Path,
		PathfinderRoot,
		AssessmentsRoot,
		endpoint)
	request = &http.Request{
		Method: method,
		Header: header,
		URL:    u,
	}
	return
}
