package auth

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

// Pages
const (
	Login           = "login"
	DeviceVerify    = "device-verify"
	DeviceSucceeded = "device-success"
	SessionExpired  = "session-expired"
)

const (
	Route = "/frontend/auth"
)

var Settings = &settings.Settings

//go:embed content/dist
var content embed.FS

var dist fs.FS

func init() {
	var err error
	if Settings.Frontend.AuthDist != "" {
		dist = os.DirFS(Settings.Frontend.AuthDist)
	} else {
		dist, err = fs.Sub(content, "content/dist")
		if err != nil {
			panic(err)
		}
	}
}

// Request defines the per-request dynamic values injected into the login page.
// Branding (titles, logos, background images) is baked into the bundle at
// build time and is NOT part of this request.
type Request struct {
	// Page selects which React component to render.
	Page string `json:"page"`

	// FormAction is the POST target URL for the login form (login page only).
	FormAction string `json:"formAction,omitempty"`

	// ErrorMessage is shown on failed authentication (login page only).
	ErrorMessage string `json:"errorMessage,omitempty"`

	// FederatedIdp renders an external IdP button when configured and not primary.
	FederatedIdp *FedIdp `json:"federatedIdp,omitempty"`

	// DeviceFormAction is the POST target URL for the device code form.
	DeviceFormAction string `json:"deviceFormAction,omitempty"`
}

// FedIdp describes the external identity provider button.
type FedIdp struct {
	// Name IdP name.
	Name string `json:"name"`
	// Login URL
	LoginURL string `json:"loginUrl"`
}

// Handler is the frontend request handler.
type Handler struct {
}

// AddRoutes adds routes.
func (h Handler) AddRoutes(e *gin.Engine) {
	h2 := http.FileServer(http.FS(dist))
	h2 = http.StripPrefix(Route, h2)
	routeGroup := e.Group(Route)
	routeGroup.GET(
		"/*path",
		func(c *gin.Context) {
			h2.ServeHTTP(c.Writer, c.Request)
		})
}

// Render the embedded HTML template with the properties
// injected as window.__LOGIN_CONFIG__ and writes the result to w.
func (h Handler) Render(w http.ResponseWriter, req Request) (err error) {
	raw, err := content.ReadFile("content/dist/index.html.tmpl")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	tp, err := template.New("index").Parse(string(raw))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	b, err := json.Marshal(req)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	w.Header().Set(api.ContentType, "text/html; charset=utf-8")
	err = tp.Execute(w, string(b))
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}
