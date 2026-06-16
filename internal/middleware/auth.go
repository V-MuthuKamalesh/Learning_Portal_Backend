package middleware

import (
	"strings"

	"github.com/collegeassess/backend/pkg/jwt"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Auth validates the Bearer access token and attaches the Principal to the context.
func Auth(jwtMgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing bearer token")
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtMgr.Parse(raw)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			return
		}
		perms := make(map[string]bool, len(claims.Permissions))
		for _, p := range claims.Permissions {
			perms[p] = true
		}
		c.Set(ctxPrincipal, &Principal{
			UserID:      claims.UserID,
			Type:        claims.Type,
			CollegeID:   claims.CollegeID,
			Role:        claims.Role,
			Permissions: perms,
		})
		c.Next()
	}
}

// RequireStudent rejects non-student principals.
func RequireStudent() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p := GetPrincipal(c); p == nil || p.Type != jwt.Student {
			response.Forbidden(c, "student access only")
			return
		}
		c.Next()
	}
}

// RequireAdmin rejects non-admin principals.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p := GetPrincipal(c); p == nil || p.Type != jwt.Admin {
			response.Forbidden(c, "admin access only")
			return
		}
		c.Next()
	}
}
