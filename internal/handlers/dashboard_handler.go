package handlers

import (
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// DashboardHandler exposes admin dashboard metrics.
type DashboardHandler struct{ svc *services.DashboardService }

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/dashboard")
	g.Use(authMW, middleware.RequireAdmin())
	g.GET("/stats", middleware.RequireAny("analytics:read", "student:read"), h.stats)
}

func (h *DashboardHandler) stats(c *gin.Context) {
	stats, err := h.svc.Stats(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to load dashboard stats")
		return
	}
	response.OK(c, stats)
}
