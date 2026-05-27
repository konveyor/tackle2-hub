package auth

import (
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Request represents an authentication request.
type Request struct {
	DB       *gorm.DB
	CTX      *gin.Context
	Token    string
	Login    string
	Password string
	Method   string
}

// Authenticate authenticates the credentials presented in the request.
func (r *Request) Authenticate() (result Result, err error) {
	jwToken, err := IdP.Authenticate(r)
	if err != nil {
		Log.Info(
			"Token not authenticated.",
			"token",
			r.Token)
		return
	}
	result.Authenticated = true
	result.Scopes = IdP.Scopes(jwToken)
	result.User = IdP.User(jwToken)
	result.Subject = IdP.Subject(jwToken)
	return
}

// With populates the request from the Authorization header.
func (r *Request) With(header string) {
	part := strings.Fields(header)
	if len(part) != 2 {
		return
	}
	method := strings.ToLower(part[0])
	switch method {
	case "bearer":
		r.Token = part[1]
	case "basic":
		encoded := part[1]
		b, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return
		}
		part := strings.SplitN(string(b), ":", 2)
		if len(part) != 2 {
			return
		}
		r.Login = part[0]
		r.Password = part[1]
	}
	return
}

// Result - auth (request) result.
type Result struct {
	Authenticated bool
	Subject       string
	User          string
	Scopes        []Scope
}
