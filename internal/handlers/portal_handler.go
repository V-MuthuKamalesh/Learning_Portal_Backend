package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StudentPortalHandler exposes student-facing endpoints.
type StudentPortalHandler struct {
	attempt  *services.AttemptService
	practice *services.PracticeService
	notif    *services.NotificationService
	coding   *services.CodingService
}

func NewStudentPortalHandler(
	attempt *services.AttemptService,
	practice *services.PracticeService,
	notif *services.NotificationService,
	coding *services.CodingService,
) *StudentPortalHandler {
	return &StudentPortalHandler{attempt: attempt, practice: practice, notif: notif, coding: coding}
}

func (h *StudentPortalHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	me := rg.Group("/me")
	me.Use(authMW, middleware.RequireStudent())
	me.GET("/assessments", h.myAssessments)
	me.GET("/results", h.myResults)
	me.GET("/notifications", h.myNotifications)
	me.PATCH("/notifications/:id/read", h.readNotification)
	me.GET("/progress", h.myProgress)

	rg.POST("/assessments/:id/start", authMW, middleware.RequireStudent(), h.start)
	rg.GET("/attempts/:id", authMW, middleware.RequireStudent(), h.getAttempt)
	rg.PUT("/attempts/:id/answers", authMW, middleware.RequireStudent(), h.saveAnswer)
	rg.POST("/attempts/:id/submit", authMW, middleware.RequireStudent(), h.submit)
	rg.POST("/attempts/:id/questions/:qid/run", authMW, middleware.RequireStudent(), h.runCoding)
	rg.POST("/attempts/:id/questions/:qid/submit", authMW, middleware.RequireStudent(), h.submitCoding)
	rg.POST("/code/run", authMW, middleware.RequireStudent(), h.runCode)

	practice := rg.Group("/practice")
	practice.Use(authMW, middleware.RequireStudent())
	practice.GET("/modules", h.practiceModules)
	practice.GET("/modules/:id", h.practiceModule)

	rg.GET("/leaderboard", authMW, middleware.RequireStudent(), h.leaderboard)
}

func (h *StudentPortalHandler) principal(c *gin.Context) (*middleware.Principal, bool) {
	p := middleware.GetPrincipal(c)
	if p == nil {
		response.Unauthorized(c, "not authenticated")
		return nil, false
	}
	return p, true
}

func (h *StudentPortalHandler) myAssessments(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	items, err := h.attempt.ListAssessments(p.CollegeID, p.UserID)
	if err != nil {
		response.Internal(c, "failed to list assessments")
		return
	}
	response.OK(c, items)
}

func (h *StudentPortalHandler) start(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	aid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	detail, err := h.attempt.Start(p.CollegeID, p.UserID, aid)
	if errors.Is(err, services.ErrAlreadySubmitted) {
		response.Conflict(c, err.Error())
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, detail)
}

func (h *StudentPortalHandler) getAttempt(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	sid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	detail, err := h.attempt.GetAttempt(sid, p.UserID)
	if notFound(c, err, "attempt") {
		return
	}
	response.OK(c, detail)
}

func (h *StudentPortalHandler) saveAnswer(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	sid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	var req dto.SaveAnswerRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.attempt.SaveAnswer(sid, p.UserID, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"saved": true})
}

func (h *StudentPortalHandler) submit(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	sid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	result, err := h.attempt.Submit(sid, p.UserID)
	if errors.Is(err, services.ErrAlreadySubmitted) {
		response.Conflict(c, err.Error())
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *StudentPortalHandler) runCode(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	var req dto.RunCodeRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	_ = p
	result, err := h.coding.RunCode(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *StudentPortalHandler) runCoding(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	sid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	qid, ok := paramUUID(c, "qid")
	if !ok {
		return
	}
	var req dto.SubmitCodingRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	result, err := h.coding.RunAttemptQuestion(sid, p.UserID, qid, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *StudentPortalHandler) submitCoding(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	sid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	qid, ok := paramUUID(c, "qid")
	if !ok {
		return
	}
	var req dto.SubmitCodingRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	result, err := h.coding.SubmitAttemptQuestion(sid, p.UserID, qid, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *StudentPortalHandler) myResults(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	items, err := h.attempt.StudentResults(p.UserID)
	if err != nil {
		response.Internal(c, "failed to load results")
		return
	}
	response.OK(c, items)
}

func (h *StudentPortalHandler) leaderboard(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	items, err := h.attempt.Leaderboard(p.CollegeID)
	if err != nil {
		response.Internal(c, "failed to load leaderboard")
		return
	}
	response.OK(c, items)
}

func (h *StudentPortalHandler) practiceModules(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	items, err := h.practice.ListModules(p.CollegeID)
	if err != nil {
		response.Internal(c, "failed to list modules")
		return
	}
	response.OK(c, items)
}

func (h *StudentPortalHandler) practiceModule(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	mid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	item, err := h.practice.ModuleDetail(p.CollegeID, mid)
	if notFound(c, err, "module") {
		return
	}
	response.OK(c, item)
}

func (h *StudentPortalHandler) myProgress(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	items, err := h.practice.StudentProgress(p.UserID)
	if err != nil {
		response.Internal(c, "failed to load progress")
		return
	}
	response.OK(c, items)
}

func (h *StudentPortalHandler) myNotifications(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	items, err := h.notif.List(p.UserID, "student")
	if err != nil {
		response.Internal(c, "failed to load notifications")
		return
	}
	response.OK(c, items)
}

func (h *StudentPortalHandler) readNotification(c *gin.Context) {
	p, ok := h.principal(c)
	if !ok {
		return
	}
	nid, ok := paramUUID(c, "id")
	if !ok {
		return
	}
	if err := h.notif.MarkRead(nid, p.UserID); err != nil {
		response.Internal(c, "failed to mark read")
		return
	}
	response.OK(c, gin.H{"is_read": true})
}

// ResultHandler exposes admin result views and CSV export.
type ResultHandler struct{ attempt *services.AttemptService }

func NewResultHandler(attempt *services.AttemptService) *ResultHandler {
	return &ResultHandler{attempt: attempt}
}

func (h *ResultHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/results")
	g.Use(authMW, middleware.RequireAdmin(), middleware.RequirePermission("result", "read"))
	g.GET("", h.list)
	g.GET("/export", middleware.RequirePermission("result", "export"), h.exportCSV)
}

func (h *ResultHandler) list(c *gin.Context) {
	var assessmentID *uuid.UUID
	if raw := c.Query("assessment_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "invalid assessment_id")
			return
		}
		assessmentID = &id
	}
	items, err := h.attempt.AdminResults(collegeScope(c), assessmentID)
	if err != nil {
		response.Internal(c, "failed to list results")
		return
	}
	response.OK(c, items)
}

func (h *ResultHandler) exportCSV(c *gin.Context) {
	items, err := h.attempt.AdminResults(collegeScope(c), nil)
	if err != nil {
		response.Internal(c, "failed to export")
		return
	}
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="results.csv"`)
	c.String(http.StatusOK, "student,assessment,marks_scored,total_marks,percentage,rank\n")
	for _, row := range items {
		name := ""
		if row.Student != nil {
			name = row.Student.Name
		}
		c.String(http.StatusOK, "%s,%s,%.1f,%d,%.1f,%d\n",
			name, row.AssessmentID, row.MarksScored, row.TotalMarks, row.Percentage, row.Rank)
	}
}

// AnalyticsHandler returns simple analytics aggregates.
type AnalyticsHandler struct {
	dashboardSvc *services.DashboardService
	attempt      *services.AttemptService
}

func NewAnalyticsHandler(dashboardSvc *services.DashboardService, attempt *services.AttemptService) *AnalyticsHandler {
	return &AnalyticsHandler{dashboardSvc: dashboardSvc, attempt: attempt}
}

func (h *AnalyticsHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/analytics")
	g.Use(authMW, middleware.RequireAdmin(), middleware.RequirePermission("analytics", "read"))
	g.GET("/dashboard", h.dashboardStats)
	g.GET("/results-summary", h.resultsSummary)
}

func (h *AnalyticsHandler) dashboardStats(c *gin.Context) {
	stats, err := h.dashboardSvc.Stats(collegeScope(c))
	if err != nil {
		response.Internal(c, "failed to load analytics")
		return
	}
	results, _ := h.attempt.AdminResults(collegeScope(c), nil)
	avg := 0.0
	if len(results) > 0 {
		sum := 0.0
		for _, r := range results {
			sum += r.Percentage
		}
		avg = sum / float64(len(results))
	}
	response.OK(c, gin.H{
		"stats":           stats,
		"average_score":   avg,
		"results_count":   len(results),
		"active_assessments": stats.Assessments,
	})
}

func (h *AnalyticsHandler) resultsSummary(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, err := h.attempt.AdminResults(collegeScope(c), nil)
	if err != nil {
		response.Internal(c, "failed to load summary")
		return
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	response.OK(c, items)
}
