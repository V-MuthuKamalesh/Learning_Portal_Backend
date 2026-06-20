package handlers

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AdminPracticeHandler manages practice modules from the admin side.
type AdminPracticeHandler struct{ svc *services.AdminPracticeService }

func NewAdminPracticeHandler(svc *services.AdminPracticeService) *AdminPracticeHandler {
	return &AdminPracticeHandler{svc: svc}
}

func (h *AdminPracticeHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/practice/modules")
	g.Use(authMW, middleware.RequireAdmin())

	g.GET("", middleware.RequirePermission("question", "read"), h.list)
	g.POST("", middleware.RequirePermission("question", "create"), h.create)
	g.GET("/:id", middleware.RequirePermission("question", "read"), h.get)
	g.PUT("/:id", middleware.RequirePermission("question", "update"), h.update)
	g.DELETE("/:id", middleware.RequirePermission("question", "delete"), h.delete)

	g.GET("/:id/questions", middleware.RequirePermission("question", "read"), h.listQuestions)
	g.POST("/:id/questions", middleware.RequirePermission("question", "update"), h.addQuestions)
	g.PATCH("/:id/questions/:qid", middleware.RequirePermission("question", "update"), h.updateQuestionSlot)
	g.DELETE("/:id/questions/:qid", middleware.RequirePermission("question", "update"), h.removeQuestion)
	g.PUT("/:id/questions/reorder", middleware.RequirePermission("question", "update"), h.reorder)
}

func (h *AdminPracticeHandler) list(c *gin.Context) {
	mods, err := h.svc.ListModules(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list modules")
		return
	}
	response.OK(c, mods)
}

func (h *AdminPracticeHandler) create(c *gin.Context) {
	var req dto.CreatePracticeModuleRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	m, err := h.svc.CreateModule(collegeScope(c), req)
	if err != nil {
		response.Internal(c, "failed to create module")
		return
	}
	response.Created(c, m)
}

func (h *AdminPracticeHandler) get(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.GetModule(collegeScope(c), id)
	if notFound(c, err, "module") {
		return
	}
	response.OK(c, m)
}

func (h *AdminPracticeHandler) update(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdatePracticeModuleRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	m, err := h.svc.UpdateModule(collegeScope(c), id, req)
	if notFound(c, err, "module") {
		return
	}
	response.OK(c, m)
}

func (h *AdminPracticeHandler) delete(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteModule(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete module")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *AdminPracticeHandler) listQuestions(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.ListModuleQuestions(collegeScope(c), id)
	if err != nil {
		response.Internal(c, "failed to list questions")
		return
	}
	response.OK(c, items)
}

func (h *AdminPracticeHandler) addQuestions(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AddModuleQuestionsRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	added, errs := h.svc.AddQuestions(collegeScope(c), id, req)
	response.OK(c, gin.H{"added": added, "failed": len(errs), "errors": errs})
}

func (h *AdminPracticeHandler) updateQuestionSlot(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	qid := c.Param("qid")
	var req dto.UpdateModuleQuestionRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.UpdateQuestionSlot(collegeScope(c), id, qid, req); notFound(c, err, "slot") {
		return
	}
	response.OK(c, gin.H{"updated": true})
}

func (h *AdminPracticeHandler) removeQuestion(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	qid := c.Param("qid")
	if err := h.svc.RemoveQuestion(collegeScope(c), id, qid); err != nil {
		response.Internal(c, "failed to remove question")
		return
	}
	response.OK(c, gin.H{"removed": true})
}

func (h *AdminPracticeHandler) reorder(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.ReorderModuleQuestionsRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.ReorderQuestions(collegeScope(c), id, req); err != nil {
		response.Internal(c, "failed to reorder")
		return
	}
	response.OK(c, gin.H{"reordered": true})
}
