package handlers

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GroupHandler exposes tenant-scoped group management endpoints.
type GroupHandler struct{ svc *services.GroupService }

func NewGroupHandler(svc *services.GroupService) *GroupHandler { return &GroupHandler{svc: svc} }

func (h *GroupHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/groups")
	g.Use(authMW, middleware.RequireAdmin())
	g.GET("", middleware.RequirePermission("group", "read"), h.list)
	g.POST("", middleware.RequirePermission("group", "create"), h.create)
	g.PUT("/:id", middleware.RequirePermission("group", "update"), h.update)
	g.DELETE("/:id", middleware.RequirePermission("group", "delete"), h.remove)
	g.GET("/:id/members", middleware.RequirePermission("group", "read"), h.listMembers)
	g.POST("/:id/members", middleware.RequirePermission("group", "update"), h.addMembers)
	g.DELETE("/:id/members/:studentId", middleware.RequirePermission("group", "update"), h.removeMember)
}

func (h *GroupHandler) list(c *gin.Context) {
	items, err := h.svc.List(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list groups")
		return
	}
	response.OK(c, items)
}

func (h *GroupHandler) create(c *gin.Context) {
	var req dto.CreateGroupRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	item, err := h.svc.Create(collegeScope(c), req)
	if err != nil {
		response.Conflict(c, "could not create group")
		return
	}
	response.Created(c, item)
}

func (h *GroupHandler) update(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateGroupRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	item, err := h.svc.Update(collegeScope(c), id, req)
	if notFound(c, err, "group") {
		return
	}
	response.OK(c, item)
}

func (h *GroupHandler) remove(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete group")
		return
	}
	response.NoContent(c)
}

func (h *GroupHandler) listMembers(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.ListMembers(collegeScope(c), id)
	if notFound(c, err, "group") {
		return
	}
	if err != nil {
		response.Internal(c, "failed to list members")
		return
	}
	response.OK(c, items)
}

func (h *GroupHandler) addMembers(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AddGroupMembersRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.StudentIDs))
	for _, raw := range req.StudentIDs {
		sid, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "invalid student_id", nil)
			return
		}
		ids = append(ids, sid)
	}
	added, err := h.svc.AddMembers(collegeScope(c), id, ids)
	if notFound(c, err, "group") {
		return
	}
	if err != nil {
		response.Internal(c, "failed to add members")
		return
	}
	response.OK(c, gin.H{"added": added})
}

func (h *GroupHandler) removeMember(c *gin.Context) {
	groupID, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	studentID, ok := paramUUID(c, "studentId")
	if !ok {
		return
	}
	if err := h.svc.RemoveMember(collegeScope(c), groupID, studentID); notFound(c, err, "group") {
		return
	}
	response.NoContent(c)
}
