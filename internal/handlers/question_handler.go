package handlers

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// QuestionHandler exposes question bank endpoints.
type QuestionHandler struct{ svc *services.QuestionService }

func NewQuestionHandler(svc *services.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: svc}
}

func (h *QuestionHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/questions")
	g.Use(authMW, middleware.RequireAdmin())
	g.GET("", middleware.RequirePermission("question", "read"), h.list)
	g.POST("/mcq", middleware.RequirePermission("question", "create"), h.createMCQ)
	g.POST("/programming", middleware.RequirePermission("question", "create"), h.createProgramming)
	g.POST("/bulk/mcq", middleware.RequirePermission("question", "create"), h.bulkMCQ)
	g.POST("/bulk/coding", middleware.RequirePermission("question", "create"), h.bulkCoding)
	g.GET("/:id", middleware.RequirePermission("question", "read"), h.get)
}

func (h *QuestionHandler) list(c *gin.Context) {
	items, err := h.svc.List(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list questions")
		return
	}
	response.OK(c, items)
}

func (h *QuestionHandler) createMCQ(c *gin.Context) {
	var req dto.CreateMCQQuestionRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	p := middleware.GetPrincipal(c)
	item, err := h.svc.CreateMCQ(collegeScope(c), p.UserID, req)
	if err != nil {
		response.Internal(c, "failed to create question")
		return
	}
	response.Created(c, item)
}

func (h *QuestionHandler) createProgramming(c *gin.Context) {
	var req dto.CreateProgrammingQuestionRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	p := middleware.GetPrincipal(c)
	item, err := h.svc.CreateProgramming(collegeScope(c), p.UserID, req)
	if err != nil {
		response.Internal(c, "failed to create question")
		return
	}
	response.Created(c, item)
}

func (h *QuestionHandler) bulkMCQ(c *gin.Context) {
	var req dto.BulkMCQRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	p := middleware.GetPrincipal(c)
	created, errs := h.svc.BulkCreateMCQ(collegeScope(c), p.UserID, req.Questions)
	response.OK(c, dto.BulkImportResult{Created: created, Failed: len(errs), Errors: errs})
}

func (h *QuestionHandler) bulkCoding(c *gin.Context) {
	var req dto.BulkCodingRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	p := middleware.GetPrincipal(c)
	created, errs := h.svc.BulkCreateProgramming(collegeScope(c), p.UserID, req.Questions)
	response.OK(c, dto.BulkImportResult{Created: created, Failed: len(errs), Errors: errs})
}

func (h *QuestionHandler) get(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.Get(collegeScope(c), id)
	if notFound(c, err, "question") {
		return
	}
	response.OK(c, item)
}
