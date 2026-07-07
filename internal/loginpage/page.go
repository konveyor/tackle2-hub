package loginpage

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/settings"
)

var Log = logr.New("loginpage", settings.Settings.Log.Auth)

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
// It is nil when the login page is not configured (assets not present).
var pageTmpl *template.Template

// tmplMu protects pageTmpl and tmplModTime from concurrent access.
var tmplMu sync.RWMutex

// tmplModTime records the modification time of the template file when it was
// last loaded so that changes can be detected on subsequent requests.
var tmplModTime time.Time

// tmplPath returns the absolute path to the login page template file.
func tmplPath() (path string) {
	path = filepath.Join(settings.Settings.LoginPage.Path, "index.html.tmpl")
	return
}

// Setup reads and parses the HTML template from the configured LoginPage.Path.
// If the template file is not found, the error is logged and pageTmpl is left
// nil so that ServeHTML returns a self-describing error page instead of
// panicking. Setup must be called once at hub startup.
func Setup() {
	path := tmplPath()
	info, err := os.Stat(path)
	if err != nil {
		Log.Error(err, "Login page template not found; OIDC login page will not be available.", "path", path)
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		Log.Error(err, "Login page template not found; OIDC login page will not be available.", "path", path)
		return
	}
	parsed, err := template.New("index").Parse(string(raw))
	if err != nil {
		Log.Error(err, "Failed to parse login page template.", "path", path)
		return
	}
	tmplMu.Lock()
	defer tmplMu.Unlock()
	pageTmpl = parsed
	tmplModTime = info.ModTime()
}

// reload checks whether the template file has been modified since last loaded
// and reparses it when a change is detected.
func reload() {
	path := tmplPath()
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	tmplMu.RLock()
	unchanged := info.ModTime().Equal(tmplModTime)
	tmplMu.RUnlock()
	if unchanged {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		Log.Error(err, "Failed to read login page template on reload.", "path", path)
		return
	}
	parsed, err := template.New("index").Parse(string(raw))
	if err != nil {
		Log.Error(err, "Failed to parse login page template on reload.", "path", path)
		return
	}
	tmplMu.Lock()
	defer tmplMu.Unlock()
	pageTmpl = parsed
	tmplModTime = info.ModTime()
	Log.Info("Login page template reloaded.", "path", path)
}

// ServeHTML executes the HTML template with the supplied config injected as
// window.__LOGIN_CONFIG__ and writes the result to w.  When the login page is
// not configured, a plain error page is written instead. The template is
// automatically reloaded when the file on disk has changed.
func ServeHTML(w http.ResponseWriter, cfg Config) (err error) {
	reload()

	tmplMu.RLock()
	tmpl := pageTmpl
	tmplMu.RUnlock()

	if tmpl == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(unconfiguredPage))
		if err != nil {
			err = liberr.Wrap(err)
		}
		return
	}

	configJSON, err := json.Marshal(cfg)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, templateData{
		ConfigJSON: string(configJSON),
	})
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// AssetHandler returns an http.Handler that serves static login page assets
// (JS, CSS, fonts, images) from the configured LoginPage.Path on disk.
// The handler strips the /oidc/assets/ prefix so that a request for
// /oidc/assets/main.abc123.js resolves to <LoginPage.Path>/main.abc123.js.
func AssetHandler() (handler http.Handler) {
	dir := http.Dir(settings.Settings.LoginPage.Path)
	handler = http.StripPrefix("/oidc/assets/", http.FileServer(dir))
	return
}

// unconfiguredPage is returned when the login page assets have not been
// installed (LOGIN_PAGE_PATH does not contain index.html.tmpl).
const unconfiguredPage = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Login Unavailable</title></head>
<body>
<h2>OIDC login page is not configured correctly.</h2>
<p>The login page assets could not be found. Contact your administrator.</p>
</body>
</html>`
