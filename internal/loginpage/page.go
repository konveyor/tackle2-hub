package loginpage

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"text/template"

	liberr "github.com/jortel/go-utils/error"
)

// Config carries the per-request dynamic values injected into the login page.
// Branding (titles, logos, background images) is baked into the bundle at
// build time and is NOT part of this config.
type Config struct {
	// Page selects which React component to render.
	Page string `json:"page"`

	// FormAction is the POST target URL for the login form (login page only).
	FormAction string `json:"formAction,omitempty"`

	// ErrorMessage is shown on failed authentication (login page only).
	ErrorMessage string `json:"errorMessage,omitempty"`

	// FederatedIdp renders an external IdP button when configured and not primary.
	FederatedIdp *FederatedIdpConfig `json:"federatedIdp,omitempty"`

	// DeviceFormAction is the POST target URL for the device code form.
	DeviceFormAction string `json:"deviceFormAction,omitempty"`
}

// FederatedIdpConfig describes the external identity provider button.
type FederatedIdpConfig struct {
	Name     string `json:"name"`
	LoginURL string `json:"loginUrl"`
}

// templateData carries values for the Go template execution.
type templateData struct {
	ConfigJSON string
}

// pageTmpl is the parsed Go template for the login page HTML.
// It is parsed once from the embedded dist/index.html.tmpl at init.
var pageTmpl *template.Template

func init() {
	raw, err := fs.ReadFile(FS, "dist/index.html.tmpl")
	if err != nil {
		panic(err)
	}
	pageTmpl = template.Must(template.New("index").Parse(string(raw)))
}

// ServeHTML executes the embedded HTML template with the supplied config
// injected as window.__LOGIN_CONFIG__ and writes the result to w.
func ServeHTML(w http.ResponseWriter, cfg Config) (err error) {
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = pageTmpl.Execute(w, templateData{
		ConfigJSON: string(configJSON),
	})
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// AssetHandler returns an http.Handler that serves static login page assets
// (JS, CSS, fonts, images) from the embedded FS.  The handler strips the
// /assets/ prefix from the request path before looking up the file so that
// a request for /oidc/assets/main.abc123.js resolves to dist/main.abc123.js.
func AssetHandler() (handler http.Handler) {
	assets, _ := fs.Sub(FS, "dist")
	handler = http.StripPrefix("/oidc/assets/", http.FileServer(http.FS(assets)))
	return
}
