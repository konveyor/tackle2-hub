package auth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/konveyor/tackle2-hub/shared/api"
)

// Issuer constructs the issuer URL from request headers.
// Checks X-Forwarded-Host, X-Forwarded-Proto, then falls back to r.Host.
func Issuer(req *http.Request) (issuer string) {
	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Host
	}

	proto := req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if req.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	issuer = fmt.Sprintf("%s://%s%s", proto, host, api.OIDCRoutes)
	return
}

// AppendIssuer appends a path to the dynamic issuer from request.
func AppendIssuer(req *http.Request, path string) (s string) {
	issuer := Issuer(req)
	issuerURL, _ := url.Parse(issuer)
	joined, _ := url.JoinPath(issuerURL.Path, path)
	issuerURL.Path = joined
	s = issuerURL.String()
	return
}
