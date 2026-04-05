package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Request authZ request.
type Request struct {
	DB     *gorm.DB
	Token  string
	Scope  string
	Method string
}

// Permit the specified request.
func (r *Request) Permit() (result Result, err error) {
	var (
		jwToken *jwt.Token
		p       Provider
	)
	for _, p = range []Provider{Hub} {
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

// Result - auth (request) result.
type Result struct {
	Authenticated bool
	Authorized    bool
	User          string
	Scopes        []Scope
}

// KeyRequest APIKey grant request.
type KeyRequest struct {
	// Userid used to authenticate a user.
	Userid string
	// Password used to authenticate the user.
	Password string
	// TaskID associated task.
	TaskID uint
	// Lifespan (TTL).
	Lifespan time.Duration
}

// Grant returns a new APIKey.
func (r KeyRequest) Grant() (key APIKey, err error) {
	key, err = Hub.Grant(r)
	return
}
