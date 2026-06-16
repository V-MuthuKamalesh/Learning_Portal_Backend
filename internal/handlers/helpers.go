package handlers

import (
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// paramUUID parses a UUID path parameter, writing a 400 and returning ok=false on failure.
func paramUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.BadRequest(c, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

// collegeScope returns the principal's college id (tenant scope for all queries).
func collegeScope(c *gin.Context) uuid.UUID {
	if p := middleware.GetPrincipal(c); p != nil {
		return p.CollegeID
	}
	return uuid.Nil
}
