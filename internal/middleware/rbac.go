package middleware

import (
	"github.com/collegeassess/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// RequirePermission enforces that the principal holds resource:action.
func RequirePermission(resource, action string) gin.HandlerFunc {
	slug := resource + ":" + action
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if !p.Has(slug) {
			response.Forbidden(c, "missing permission: "+slug)
			return
		}
		c.Next()
	}
}

// RequireAny passes if the principal holds at least one of the given permission slugs.
func RequireAny(slugs ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		for _, s := range slugs {
			if p.Has(s) {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "insufficient permissions")
	}
}
