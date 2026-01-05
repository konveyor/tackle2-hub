package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Request auth request.
type Request struct {
	Token  string
	Scope  string
	Method string
	DB     *gorm.DB
}

// Permit the specified request.
func (r *Request) Permit() (result Result, err error) {
	var (
		jwToken *jwt.Token
		p       Provider
	)
	for _, p = range []Provider{Hub, Remote} {
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
	if err != nil {
		return
	}
	if result.Authenticated {
		scopes := p.Scopes(jwToken)
		for _, scope := range scopes {
			if scope.Match(r.Scope, r.Method) {
				result.Scopes = scopes
				result.User = p.User(jwToken)
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

// Result - auth result.
type Result struct {
	Authenticated bool
	Authorized    bool
	User          string
	Scopes        []Scope
}
