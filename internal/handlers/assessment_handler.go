package handlers

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AssessmentHandler exposes assessment management endpoints.
type AssessmentHandler struct {
	svc     *services.AssessmentService
	attempt *services.AttemptService
}

func NewAssessmentHandler(svc *services.AssessmentService, attempt *services.AttemptService) *AssessmentHandler {
	return &AssessmentHandler{svc: svc, attempt: attempt}
}

func (h *AssessmentHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/assessments")
	g.Use(authMW, middleware.RequireAdmin())
	g.GET("", middleware.RequirePermission("assessment", "read"), h.list)
	g.POST("", middleware.RequirePermission("assessment", "create"), h.create)
	g.GET("/:id", middleware.RequirePermission("assessment", "read"), h.get)
	g.PUT("/:id", middleware.RequirePermission("assessment", "update"), h.update)
	g.DELETE("/:id", middleware.RequirePermission("assessment", "delete"), h.remove)
	g.POST("/:id/questions", middleware.RequirePermission("assessment", "update"), h.attachQuestions)
	g.POST("/:id/assign", middleware.RequirePermission("assessment", "update"), h.assign)
	g.POST("/:id/publish", middleware.RequirePermission("assessment", "publish"), h.publish)
	g.GET("/:id/results", middleware.RequirePermission("result", "read"), h.results)
}

func (h *AssessmentHandler) list(c *gin.Context) {
	items, err := h.svc.List(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list assessments")
		return
	}
	response.OK(c, items)
}

func (h *AssessmentHandler) create(c *gin.Context) {
	var req dto.CreateAssessmentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	p := middleware.GetPrincipal(c)
	item, err := h.svc.Create(collegeScope(c), p.UserID, req)
	if err != nil {
		response.Internal(c, "failed to create assessment")
		return
	}
	response.Created(c, item)
}

func (h *AssessmentHandler) get(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.Get(collegeScope(c), id)
	if notFound(c, err, "assessment") {
		return
	}
	response.OK(c, item)
}

func (h *AssessmentHandler) update(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAssessmentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	item, err := h.svc.Update(collegeScope(c), id, req)
	if notFound(c, err, "assessment") {
		return
	}
	response.OK(c, item)
}

func (h *AssessmentHandler) remove(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete assessment")
		return
	}
	response.NoContent(c)
}

func (h *AssessmentHandler) attachQuestions(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AttachQuestionsRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.AttachQuestions(collegeScope(c), id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"attached": len(req.QuestionIDs)})
}

func (h *AssessmentHandler) assign(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.AssignAssessmentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.Assign(collegeScope(c), id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"assigned": true})
}

func (h *AssessmentHandler) publish(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.Publish(collegeScope(c), id)
	if notFound(c, err, "assessment") {
		return
	}
	response.OK(c, item)
}

func (h *AssessmentHandler) results(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	items, err := h.attempt.AdminResults(collegeScope(c), &id)
	if err != nil {
		response.Internal(c, "failed to list results")
		return
	}
	response.OK(c, items)
}
