package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/settings"
	"net/http"
	"strings"
)

//
// AuthorizationRequired enforces that the user (identified by a token) has
// been granted the necessary scope to access a resource.
func AuthorizationRequired(p Provider, requiredScope string) func(*gin.Context) {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		secret := settings.Settings.Auth.AddonAccessSecret
		if secret != "" && token == secret {
			c.Next()
			return
		}

		scopes, err := p.Scopes(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		for _, s := range scopes {
			if s.Allow(requiredScope, strings.ToLower(c.Request.Method)) {
				c.Next()
				return
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}
