package loginpage

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

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

// configPlaceholder is the exact string emitted by rspack's minifier for the
// placeholder in src/index.html.  The hub replaces this with the real config
// before writing the response.
const configPlaceholder = "window.__LOGIN_CONFIG__=null"

// ServeHTML reads index.html from the embedded FS, injects the supplied config
// as window.__LOGIN_CONFIG__, and writes the result to w.
func ServeHTML(w http.ResponseWriter, cfg Config) (err error) {
	raw, err := fs.ReadFile(FS, "dist/index.html")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	configJSON, err := json.Marshal(cfg)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	injected := "window.__LOGIN_CONFIG__=" + string(configJSON)
	html := strings.Replace(string(raw), configPlaceholder, injected, 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = w.Write([]byte(html))
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

// AssetHandler returns an http.Handler that serves static login page assets
// (JS, CSS, fonts, images) from the embedded FS.  The handler strips the
// /assets/ prefix from the request path before looking up the file so that
// a request for /oidc/assets/main.abc123.js resolves to dist/main.abc123.js.
func AssetHandler() http.Handler {
	assets, _ := fs.Sub(FS, "dist")
	return http.StripPrefix("/oidc/assets/", http.FileServer(http.FS(assets)))
}
