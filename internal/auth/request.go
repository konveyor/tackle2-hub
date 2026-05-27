package auth

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Request represents an authentication and authorization request.
type Request struct {
	DB       *gorm.DB
	CTX      *gin.Context
	Token    string
	Login    string
	Password string
	Scope    string
	Method   string
}

// Permit authenticates and authorizes the request.
func (r *Request) Permit() (result Result, err error) {
	var (
		jwToken *jwt.Token
		p       Provider
	)
	for _, p = range []Provider{IdP} {
		var pErr error
		jwToken, pErr = p.Authenticate(r)
		if pErr == nil {
			result.Authenticated = true
			break
		}
		if errors.Is(pErr, &NotAuthenticated{}) {
			continue
		}
		if errors.Is(pErr, &NotValid{}) {
			break
		}
		err = pErr
		return

	}
	if result.Authenticated {
		scopes := p.Scopes(jwToken)
		for _, scope := range scopes {
			if r.Scope == "" || scope.Match(r.Scope, r.Method) {
				result.Scopes = scopes
				result.User = p.User(jwToken)
				result.Subject = p.Subject(jwToken)
				result.Authorized = true
				break
			}
		}
	} else {
		Log.Info(
			"Token not authenticated.",
			"token",
			r.Token)
	}
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
	Authorized    bool
	Subject       string
	User          string
	Scopes        []Scope
}
