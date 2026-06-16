package handlers

import (
	"errors"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// CollegeHandler exposes tenant, branding, department and batch endpoints.
type CollegeHandler struct{ svc *services.CollegeService }

func NewCollegeHandler(svc *services.CollegeService) *CollegeHandler {
	return &CollegeHandler{svc: svc}
}

// RegisterPublic mounts unauthenticated branding lookup (for the login page).
func (h *CollegeHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.GET("/colleges/:id/public", h.branding)
}

// Register mounts authenticated college/department/batch routes.
func (h *CollegeHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("")
	g.Use(authMW, middleware.RequireAdmin())

	// College management (super admin: college:manage).
	manage := middleware.RequirePermission("college", "manage")
	g.GET("/colleges", manage, h.list)
	g.POST("/colleges", manage, h.create)
	g.GET("/colleges/:id", middleware.RequireAny("college:read", "college:manage"), h.get)
	g.PUT("/colleges/:id", manage, h.update)
	g.DELETE("/colleges/:id", manage, h.remove)
	g.POST("/colleges/:id/logo", manage, h.uploadLogo)

	// Departments & batches (tenant-scoped).
	g.GET("/departments", h.listDepartments)
	g.POST("/departments", manage, h.createDepartment)
	g.PUT("/departments/:id", manage, h.updateDepartment)
	g.DELETE("/departments/:id", manage, h.deleteDepartment)

	g.GET("/batches", h.listBatches)
	g.POST("/batches", manage, h.createBatch)
	g.PUT("/batches/:id", manage, h.updateBatch)
	g.DELETE("/batches/:id", manage, h.deleteBatch)
}

func (h *CollegeHandler) branding(c *gin.Context) {
	b, err := h.svc.Branding(c.Param("id"))
	if err != nil {
		response.NotFound(c, "college not found")
		return
	}
	response.OK(c, b)
}

func (h *CollegeHandler) list(c *gin.Context) {
	cs, err := h.svc.List()
	if err != nil {
		response.Internal(c, "failed to list colleges")
		return
	}
	response.OK(c, cs)
}

func (h *CollegeHandler) create(c *gin.Context) {
	var req dto.CreateCollegeRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	col, err := h.svc.Create(req)
	if err != nil {
		response.Conflict(c, "could not create college (code may already exist)")
		return
	}
	response.Created(c, col)
}

func (h *CollegeHandler) get(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	col, err := h.svc.Get(id)
	if notFound(c, err, "college") {
		return
	}
	response.OK(c, col)
}

func (h *CollegeHandler) update(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateCollegeRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	col, err := h.svc.Update(id, req)
	if notFound(c, err, "college") {
		return
	}
	response.OK(c, col)
}

func (h *CollegeHandler) remove(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Internal(c, "failed to delete college")
		return
	}
	response.NoContent(c)
}

func (h *CollegeHandler) uploadLogo(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
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
	url, err := h.svc.UploadLogo(id, file.Filename, f)
	if err != nil {
		response.Internal(c, "upload failed: "+err.Error())
		return
	}
	response.OK(c, gin.H{"logo_url": url})
}

// ── Departments ──
func (h *CollegeHandler) listDepartments(c *gin.Context) {
	ds, err := h.svc.ListDepartments(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list departments")
		return
	}
	response.OK(c, ds)
}

func (h *CollegeHandler) createDepartment(c *gin.Context) {
	var req dto.DepartmentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	d, err := h.svc.CreateDepartment(collegeScope(c), req)
	if err != nil {
		response.Conflict(c, "could not create department")
		return
	}
	response.Created(c, d)
}

func (h *CollegeHandler) updateDepartment(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.DepartmentRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	d, err := h.svc.UpdateDepartment(collegeScope(c), id, req)
	if notFound(c, err, "department") {
		return
	}
	response.OK(c, d)
}

func (h *CollegeHandler) deleteDepartment(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteDepartment(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete department")
		return
	}
	response.NoContent(c)
}

// ── Batches ──
func (h *CollegeHandler) listBatches(c *gin.Context) {
	bs, err := h.svc.ListBatches(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to list batches")
		return
	}
	response.OK(c, bs)
}

func (h *CollegeHandler) createBatch(c *gin.Context) {
	var req dto.BatchRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	b, err := h.svc.CreateBatch(collegeScope(c), req)
	if err != nil {
		response.Conflict(c, "could not create batch")
		return
	}
	response.Created(c, b)
}

func (h *CollegeHandler) updateBatch(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.BatchRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	b, err := h.svc.UpdateBatch(collegeScope(c), id, req)
	if notFound(c, err, "batch") {
		return
	}
	response.OK(c, b)
}

func (h *CollegeHandler) deleteBatch(c *gin.Context) {
	id, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteBatch(collegeScope(c), id); err != nil {
		response.Internal(c, "failed to delete batch")
		return
	}
	response.NoContent(c)
}

// notFound writes a 404 when err is ErrNotFound; returns true if it handled the error.
func notFound(c *gin.Context, err error, resource string) bool {
	if errors.Is(err, repositories.ErrNotFound) {
		response.NotFound(c, resource+" not found")
		return true
	}
	if err != nil {
		response.Internal(c, "unexpected error")
		return true
	}
	return false
}
