package handlers

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AdminHandler exposes admin user management endpoints.
type AdminHandler struct{ svc *services.AdminService }

func NewAdminHandler(svc *services.AdminService) *AdminHandler { return &AdminHandler{svc: svc} }

func (h *AdminHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/admins")
	g.Use(authMW, middleware.RequireAdmin())
	g.GET("", middleware.RequirePermission("admin", "read"), h.list)
	g.POST("", middleware.RequirePermission("admin", "create"), h.create)
	g.PATCH("/:id/status", middleware.RequirePermission("admin", "update"), h.status)
	g.DELETE("/:id", middleware.RequirePermission("admin", "delete"), h.remove)

	rg.GET("/roles", authMW, middleware.RequireAdmin(), middleware.RequirePermission("role", "read"), h.roles)
}

func (h *AdminHandler) list(c *gin.Context) {
	items, err := h.svc.List(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list admins")
		return
	}
	response.OK(c, items)
}

func (h *AdminHandler) create(c *gin.Context) {
	var req dto.CreateAdminRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	item, err := h.svc.Create(collegeScope(c), req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, item)
}

func (h *AdminHandler) status(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AdminStatusRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.SetStatus(collegeScope(c), id, req.IsActive); notFound(c, err, "admin") {
		return
	}
	response.OK(c, gin.H{"is_active": req.IsActive})
}

func (h *AdminHandler) remove(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete admin")
		return
	}
	response.NoContent(c)
}

func (h *AdminHandler) roles(c *gin.Context) {
	items, err := h.svc.ListRoles()
	if err != nil {
		response.Internal(c, "failed to list roles")
		return
	}
	response.OK(c, items)
}
