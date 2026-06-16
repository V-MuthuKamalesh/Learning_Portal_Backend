package dto

import "time"

// StudentLoginRequest authenticates a student.
type StudentLoginRequest struct {
	CollegeCode    string `json:"college_code"`
	RegisterNumber string `json:"register_number" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
}

// AdminLoginRequest authenticates an admin.
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RefreshRequest rotates a refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest starts a reset flow.
type ForgotPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	UserType string `json:"user_type" binding:"required,oneof=student admin"`
}

// ResetPasswordRequest completes a reset flow.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePasswordRequest changes the password of the logged-in principal.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// VerifyEmailRequest confirms an email token.
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// AuthUser is the principal returned alongside tokens.
type AuthUser struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Type        string    `json:"type"`
	CollegeID   string    `json:"college_id"`
	Role        string    `json:"role,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"`
}

// TokenPair is the access/refresh response.
type TokenPair struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"` // access TTL seconds
	User         AuthUser `json:"user"`
}
