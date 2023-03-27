package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
)

//
// Request auth request.
type Request struct {
	Token  string
	Scope  string
	Method string
}

//
// Permit the specified request.
func (r *Request) Permit() (result Result, err error) {
	var (
		jwToken *jwt.Token
		p       Provider
	)
	for _, p = range []Provider{Hub, Remote} {
		var pErr error
		jwToken, pErr = p.Authenticate(r.Token)
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
