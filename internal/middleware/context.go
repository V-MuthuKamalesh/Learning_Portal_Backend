// Package middleware holds cross-cutting HTTP concerns and the request principal.
package middleware

import (
	"github.com/collegeassess/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ctxPrincipal = "principal"
	ctxRequestID = "request_id"
)

// Principal is the authenticated identity attached to a request.
type Principal struct {
	UserID      uuid.UUID
	Type        jwt.PrincipalType
	CollegeID   uuid.UUID
	Role        string
	Permissions map[string]bool
}

// Has reports whether the principal holds a permission (honoring the "*" wildcard).
func (p *Principal) Has(slug string) bool {
	if p == nil {
		return false
	}
	return p.Permissions["*"] || p.Permissions[slug]
}

// GetPrincipal extracts the principal from the gin context.
func GetPrincipal(c *gin.Context) *Principal {
	v, ok := c.Get(ctxPrincipal)
	if !ok {
		return nil
	}
	p, _ := v.(*Principal)
	return p
}

// RequestID returns the per-request correlation id.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
