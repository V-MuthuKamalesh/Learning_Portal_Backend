package handlers

import (
	"net/http"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/pagination"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// StudentHandler exposes tenant-scoped student management endpoints.
type StudentHandler struct{ svc *services.StudentService }

func NewStudentHandler(svc *services.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

// Register mounts authenticated student management routes.
func (h *StudentHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/students")
	g.Use(authMW, middleware.RequireAdmin())

	g.GET("", middleware.RequirePermission("student", "read"), h.list)
	g.POST("", middleware.RequirePermission("student", "create"), h.create)
	g.GET("/import-template", middleware.RequirePermission("student", "read"), h.importTemplate)
	g.POST("/bulk-import", middleware.RequirePermission("student", "create"), h.bulkImport)
	g.GET("/:id", middleware.RequirePermission("student", "read"), h.get)
	g.PUT("/:id", middleware.RequirePermission("student", "update"), h.update)
	g.DELETE("/:id", middleware.RequirePermission("student", "delete"), h.remove)
	g.PATCH("/:id/status", middleware.RequirePermission("student", "update"), h.status)
}

func (h *StudentHandler) list(c *gin.Context) {
	p := pagination.Parse(c)
	filter := dto.StudentFilter{
		DepartmentID: c.Query("department_id"),
		BatchID:      c.Query("batch_id"),
		GroupID:      c.Query("group_id"),
		Year:         c.Query("year"),
		Section:      c.Query("section"),
		Status:       c.Query("status"),
	}
	students, total, err := h.svc.List(collegeScope(c), nil, filter, p)
	if err != nil {
		response.Internal(c, "failed to list students")
		return
	}
	response.List(c, students, p.Meta(total))
}

func (h *StudentHandler) create(c *gin.Context) {
	var req dto.CreateStudentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	student, err := h.svc.Create(collegeScope(c), req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, student)
}

func (h *StudentHandler) get(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	student, err := h.svc.Get(collegeScope(c), id)
	if notFound(c, err, "student") {
		return
	}
	response.OK(c, student)
}

func (h *StudentHandler) update(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateStudentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	student, err := h.svc.Update(collegeScope(c), id, req)
	if notFound(c, err, "student") {
		return
	}
	response.OK(c, student)
}

func (h *StudentHandler) remove(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete student")
		return
	}
	response.NoContent(c)
}

func (h *StudentHandler) status(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.StudentStatusRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.SetStatus(collegeScope(c), id, req.IsActive); err != nil {
		response.NotFound(c, "student not found")
		return
	}
	response.OK(c, gin.H{"is_active": req.IsActive})
}

func (h *StudentHandler) bulkImport(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file field required")
		return
	}
	f, err := file.Open()
	if err != nil {
		response.BadRequest(c, "could not read upload")
		return
	}
	defer f.Close()

	result, err := h.svc.BulkImport(collegeScope(c), f)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *StudentHandler) importTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="students-import-template.csv"`)
	c.String(http.StatusOK, "name,register_number,email,department_code,year,section,phone\n")
}
