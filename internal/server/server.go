// Package server builds the HTTP engine, wires dependencies and registers routes.
package server

import (
	"github.com/collegeassess/backend/configs"
	"github.com/collegeassess/backend/internal/handlers"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/jwt"
	"github.com/collegeassess/backend/pkg/mailer"
	"github.com/collegeassess/backend/pkg/judge"
	"github.com/collegeassess/backend/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Deps are the external resources the server needs.
type Deps struct {
	Cfg   *configs.Config
	DB    *gorm.DB
	Redis *redis.Client
}

// New constructs a fully-wired gin engine.
func New(d Deps) *gin.Engine {
	if d.Cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware(), middleware.Logger(), middleware.Recovery(), middleware.CORS(d.Cfg.Cors))

	if d.Cfg.Storage.Driver == "local" {
		r.Static("/uploads", d.Cfg.Storage.LocalDir)
	}

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	jwtMgr := jwt.NewManager(d.Cfg.JWT.AccessSecret, d.Cfg.JWT.AccessTTL)
	mail := mailer.NewLogMailer(d.Cfg.SMTP.From)
	authMW := middleware.Auth(jwtMgr)

	store := storage.New(d.Cfg)

	authRepo := repositories.NewAuthRepository(d.DB)
	authSvc := services.NewAuthService(authRepo, jwtMgr, d.Cfg, mail)
	authHandler := handlers.NewAuthHandler(authSvc)

	collegeRepo := repositories.NewCollegeRepository(d.DB)
	collegeSvc := services.NewCollegeService(collegeRepo, store)
	collegeHandler := handlers.NewCollegeHandler(collegeSvc)

	studentRepo := repositories.NewStudentRepository(d.DB)
	studentSvc := services.NewStudentService(studentRepo, collegeRepo)
	studentHandler := handlers.NewStudentHandler(studentSvc)

	groupRepo := repositories.NewGroupRepository(d.DB)
	groupSvc := services.NewGroupService(groupRepo)
	groupHandler := handlers.NewGroupHandler(groupSvc)

	questionRepo := repositories.NewQuestionRepository(d.DB)
	questionSvc := services.NewQuestionService(questionRepo)
	questionHandler := handlers.NewQuestionHandler(questionSvc)

	assessmentRepo := repositories.NewAssessmentRepository(d.DB)
	attemptRepo := repositories.NewAttemptRepository(d.DB)
	notifRepo := repositories.NewNotificationRepository(d.DB)
	practiceRepo := repositories.NewPracticeRepository(d.DB)

	assessmentSvc := services.NewAssessmentService(assessmentRepo)
	attemptSvc := services.NewAttemptService(assessmentRepo, attemptRepo, studentRepo, notifRepo)
	judgeClient := judge.NewClient(d.Cfg.Judge)
	codingSvc := services.NewCodingService(judgeClient, attemptRepo, assessmentRepo, studentRepo)
	practiceSvc := services.NewPracticeService(practiceRepo, studentRepo, attemptRepo)
	adminPracticeSvc := services.NewAdminPracticeService(practiceRepo, questionRepo)
	notifSvc := services.NewNotificationService(notifRepo)

	assessmentHandler := handlers.NewAssessmentHandler(assessmentSvc, attemptSvc)
	studentPortalHandler := handlers.NewStudentPortalHandler(attemptSvc, practiceSvc, notifSvc, codingSvc)
	adminPracticeHandler := handlers.NewAdminPracticeHandler(adminPracticeSvc)
	resultHandler := handlers.NewResultHandler(attemptSvc)

	adminRepo := repositories.NewAdminRepository(d.DB)
	roleRepo := repositories.NewRoleRepository(d.DB)
	adminSvc := services.NewAdminService(adminRepo, roleRepo)
	adminHandler := handlers.NewAdminHandler(adminSvc)

	dashboardSvc := services.NewDashboardService(studentRepo, groupRepo, assessmentRepo, questionRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc)
	analyticsHandler := handlers.NewAnalyticsHandler(dashboardSvc, attemptSvc)

	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(120))
	authHandler.Register(v1, authMW)
	collegeHandler.RegisterPublic(v1)
	collegeHandler.Register(v1, authMW)
	studentHandler.Register(v1, authMW)
	groupHandler.Register(v1, authMW)
	questionHandler.Register(v1, authMW)
	assessmentHandler.Register(v1, authMW)
	dashboardHandler.Register(v1, authMW)
	adminHandler.Register(v1, authMW)
	studentPortalHandler.Register(v1, authMW)
	resultHandler.Register(v1, authMW)
	analyticsHandler.Register(v1, authMW)
	adminPracticeHandler.Register(v1, authMW)

	RegisterSwagger(r)
	return r
}
