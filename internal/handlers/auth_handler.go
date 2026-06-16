// Package handlers is the HTTP layer: bind, delegate to services, map errors to responses.
package handlers

import (
	"errors"
	"net/http"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/middleware"
	"github.com/collegeassess/backend/internal/services"
	"github.com/collegeassess/backend/pkg/response"
	"github.com/collegeassess/backend/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct{ svc *services.AuthService }

func NewAuthHandler(svc *services.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

// Register wires auth routes under the given group.
func (h *AuthHandler) Register(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/auth")
	g.POST("/student/login", h.studentLogin)
	g.POST("/admin/login", h.adminLogin)
	g.POST("/refresh", h.refresh)
	g.POST("/forgot-password", h.forgotPassword)
	g.POST("/reset-password", h.resetPassword)
	g.POST("/verify-email", h.verifyEmail)

	authed := g.Group("")
	authed.Use(authMW)
	authed.POST("/logout", h.logout)
	authed.POST("/change-password", h.changePassword)
	authed.GET("/me", h.me)
}

// studentLogin godoc
// @Summary Student login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.StudentLoginRequest true "credentials"
// @Success 200 {object} dto.TokenPair
// @Router /auth/student/login [post]
func (h *AuthHandler) studentLogin(c *gin.Context) {
	var req dto.StudentLoginRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	pair, err := h.svc.StudentLogin(req)
	if err != nil {
		h.mapAuthError(c, err)
		return
	}
	response.OK(c, pair)
}

// adminLogin godoc
// @Summary Admin login
// @Tags auth
// @Param body body dto.AdminLoginRequest true "credentials"
// @Success 200 {object} dto.TokenPair
// @Router /auth/admin/login [post]
func (h *AuthHandler) adminLogin(c *gin.Context) {
	var req dto.AdminLoginRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	pair, err := h.svc.AdminLogin(req)
	if err != nil {
		h.mapAuthError(c, err)
		return
	}
	response.OK(c, pair)
}

func (h *AuthHandler) refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	pair, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}
	response.OK(c, pair)
}

func (h *AuthHandler) logout(c *gin.Context) {
	var req dto.RefreshRequest
	_ = validator.BindJSON(c, &req)
	_ = h.svc.Logout(req.RefreshToken)
	response.OK(c, gin.H{"message": "logged out"})
}

func (h *AuthHandler) forgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	h.svc.ForgotPassword(req)
	response.OK(c, gin.H{"message": "if the account exists, a reset link has been sent"})
}

func (h *AuthHandler) resetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.ResetPassword(req); err != nil {
		response.BadRequest(c, "invalid or expired reset token")
		return
	}
	response.OK(c, gin.H{"message": "password reset successful"})
}

func (h *AuthHandler) changePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	p := middleware.GetPrincipal(c)
	if err := h.svc.ChangePassword(p.Type, p.UserID, req); err != nil {
		response.BadRequest(c, "current password is incorrect")
		return
	}
	response.OK(c, gin.H{"message": "password changed"})
}

func (h *AuthHandler) verifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if errs := validator.BindJSON(c, &req); errs != nil {
		response.BadRequest(c, "validation failed", errs)
		return
	}
	if err := h.svc.VerifyEmail(req); err != nil {
		response.BadRequest(c, "invalid verification token")
		return
	}
	response.OK(c, gin.H{"message": "email verified"})
}

func (h *AuthHandler) me(c *gin.Context) {
	p := middleware.GetPrincipal(c)
	perms := make([]string, 0, len(p.Permissions))
	for k := range p.Permissions {
		perms = append(perms, k)
	}
	response.OK(c, gin.H{
		"id":          p.UserID,
		"type":        p.Type,
		"college_id":  p.CollegeID,
		"role":        p.Role,
		"permissions": perms,
	})
}

func (h *AuthHandler) mapAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrAccountLocked):
		response.Locked(c, "account locked due to too many failed attempts; try again later")
	case errors.Is(err, services.ErrAccountInactive):
		response.Forbidden(c, "account is deactivated")
	case errors.Is(err, services.ErrInvalidCredentials):
		response.Unauthorized(c, "invalid credentials")
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "something went wrong"}})
	}
}
