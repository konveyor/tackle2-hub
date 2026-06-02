package auth

import (
	"net/http"
	"net/url"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/vfaronov/httpheader"
)

// ReqInCtx key used to store the request in a context.
const (
	ReqInCtx = "http.request"
)

// Issuer constructs the issuer URL based on `Forward` header.
// When not present, uses the request schema and host.
// Forward RFC-7239 standard header adopted in 2014.
func Issuer(req *http.Request) (issuer string) {
	proto := "http"
	if req.TLS != nil {
		proto = "https"
	}
	host := req.Host
	forwarded := httpheader.Forwarded(req.Header)
	if len(forwarded) > 0 {
		proto = forwarded[0].Proto
		host = forwarded[0].Host
	}
	issuer = proto + "://" + host + api.OIDCRoutes
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
