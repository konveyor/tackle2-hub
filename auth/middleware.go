package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

const (
	Header      = "Authorization"
	TokenScopes = "Scopes"
	TokenUser   = "User"
)

//
// Required enforces that the user (identified by a token) has
// been granted the necessary scope to access a resource.
func Required(requiredScope string) func(*gin.Context) {
	return func(c *gin.Context) {
		var (
			matched bool
			err     error
		)
		token := c.GetHeader(Header)
		var jwToken *jwt.Token
		for _, p := range []Provider{Hub, Remote} {
			jwToken, err = p.Authenticate(token)
			if err != nil {
				if errors.Is(err, &NotAuthenticated{}) {
					continue
				}
				if errors.Is(err, &NotValid{}) {
					break
				}
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			scopes := p.Scopes(jwToken)
			for _, scope := range scopes {
				if scope.Match(requiredScope, c.Request.Method) {
					c.Set(TokenUser, p.User(jwToken))
					c.Set(TokenScopes, scopes)
					matched = true
					break
				}
			}
			break
		}
		switch {
		case errors.Is(err, &NotValid{}):
			c.AbortWithStatus(http.StatusForbidden)
		case errors.Is(err, &NotAuthenticated{}):
			c.AbortWithStatus(http.StatusUnauthorized)
		default:
			if !matched {
				c.AbortWithStatus(http.StatusForbidden)
			} else {
				c.Next()
			}
		}

		return
	}
}
