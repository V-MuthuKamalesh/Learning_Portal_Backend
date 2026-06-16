package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/collegeassess/backend/configs"
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/collegeassess/backend/pkg/hash"
	"github.com/collegeassess/backend/pkg/jwt"
	"github.com/collegeassess/backend/pkg/mailer"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Sentinel service errors mapped to HTTP codes by the handler.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked")
	ErrAccountInactive    = errors.New("account inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService implements authentication, lockout, refresh rotation and password flows.
type AuthService struct {
	repo   *repositories.AuthRepository
	jwtMgr *jwt.Manager
	cfg    *configs.Config
	mail   mailer.Mailer
}

func NewAuthService(repo *repositories.AuthRepository, jwtMgr *jwt.Manager, cfg *configs.Config, mail mailer.Mailer) *AuthService {
	return &AuthService{repo: repo, jwtMgr: jwtMgr, cfg: cfg, mail: mail}
}

// StudentLogin authenticates a student and issues a token pair.
func (s *AuthService) StudentLogin(req dto.StudentLoginRequest) (*dto.TokenPair, error) {
	var collegeID *uuid.UUID
	if req.CollegeCode != "" {
		col, err := s.repo.FindCollegeByCode(req.CollegeCode)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		collegeID = &col.ID
	}

	student, err := s.repo.FindStudent(collegeID, req.RegisterNumber, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if locked, _ := s.isLocked(student.LockedUntil); locked {
		return nil, ErrAccountLocked
	}
	if !hash.CheckPassword(student.PasswordHash, req.Password) {
		s.registerFailure(&student.FailedAttempts, &student.LockedUntil)
		_ = s.repo.UpdateStudent(student)
		return nil, ErrInvalidCredentials
	}
	if !student.IsActive {
		return nil, ErrAccountInactive
	}

	now := time.Now()
	student.FailedAttempts = 0
	student.LockedUntil = nil
	student.LastLoginAt = &now
	_ = s.repo.UpdateStudent(student)

	claims := jwt.Claims{UserID: student.ID, Type: jwt.Student, CollegeID: student.CollegeID}
	user := dto.AuthUser{
		ID: student.ID.String(), Name: student.Name, Email: student.Email,
		Type: "student", CollegeID: student.CollegeID.String(), LastLoginAt: now,
	}
	return s.issueTokens(claims, jwt.Student, user)
}

// AdminLogin authenticates an admin (with lockout) and issues a token pair.
func (s *AuthService) AdminLogin(req dto.AdminLoginRequest) (*dto.TokenPair, error) {
	admin, err := s.repo.FindAdminByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if locked, _ := s.isLocked(admin.LockedUntil); locked {
		return nil, ErrAccountLocked
	}
	if !hash.CheckPassword(admin.PasswordHash, req.Password) {
		s.registerFailure(&admin.FailedAttempts, &admin.LockedUntil)
		_ = s.repo.UpdateAdmin(admin)
		return nil, ErrInvalidCredentials
	}
	if !admin.IsActive {
		return nil, ErrAccountInactive
	}

	now := time.Now()
	admin.FailedAttempts = 0
	admin.LockedUntil = nil
	admin.LastLoginAt = &now
	_ = s.repo.UpdateAdmin(admin)

	perms := []string{}
	roleName := ""
	if admin.Role != nil {
		perms = admin.Role.PermissionSlugs()
		roleName = admin.Role.Name
	}
	claims := jwt.Claims{
		UserID: admin.ID, Type: jwt.Admin, CollegeID: admin.CollegeID,
		Role: roleName, Permissions: perms,
	}
	user := dto.AuthUser{
		ID: admin.ID.String(), Name: admin.Name, Email: admin.Email, Type: "admin",
		CollegeID: admin.CollegeID.String(), Role: roleName, Permissions: perms, LastLoginAt: now,
	}
	return s.issueTokens(claims, jwt.Admin, user)
}

// Refresh validates an opaque refresh token, rotates it and returns a new pair.
func (s *AuthService) Refresh(raw string) (*dto.TokenPair, error) {
	stored, err := s.repo.FindRefreshToken(hash.SHA256(raw))
	if err != nil {
		return nil, ErrInvalidToken
	}
	if stored.Revoked {
		// Token reuse → revoke the entire family.
		_ = s.repo.RevokeFamily(stored.FamilyID)
		return nil, ErrInvalidToken
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidToken
	}
	_ = s.repo.RevokeRefreshToken(stored.ID)

	// Rebuild claims from the current identity.
	var claims jwt.Claims
	var user dto.AuthUser
	if stored.UserType == string(jwt.Admin) {
		admin, err := s.repo.FindAdminByID(stored.UserID)
		if err != nil {
			return nil, ErrInvalidToken
		}
		perms, roleName := []string{}, ""
		if admin.Role != nil {
			perms, roleName = admin.Role.PermissionSlugs(), admin.Role.Name
		}
		claims = jwt.Claims{UserID: admin.ID, Type: jwt.Admin, CollegeID: admin.CollegeID, Role: roleName, Permissions: perms}
		user = dto.AuthUser{ID: admin.ID.String(), Name: admin.Name, Email: admin.Email, Type: "admin", CollegeID: admin.CollegeID.String(), Role: roleName, Permissions: perms}
	} else {
		student, err := s.repo.FindStudentByID(stored.UserID)
		if err != nil {
			return nil, ErrInvalidToken
		}
		claims = jwt.Claims{UserID: student.ID, Type: jwt.Student, CollegeID: student.CollegeID}
		user = dto.AuthUser{ID: student.ID.String(), Name: student.Name, Email: student.Email, Type: "student", CollegeID: student.CollegeID.String()}
	}

	return s.issueTokensWithFamily(claims, jwt.PrincipalType(stored.UserType), user, stored.FamilyID)
}

// Logout revokes a single refresh token.
func (s *AuthService) Logout(raw string) error {
	stored, err := s.repo.FindRefreshToken(hash.SHA256(raw))
	if err != nil {
		return nil
	}
	return s.repo.RevokeRefreshToken(stored.ID)
}

// ChangePassword verifies the old password and sets a new one.
func (s *AuthService) ChangePassword(userType jwt.PrincipalType, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	if userType == jwt.Admin {
		admin, err := s.repo.FindAdminByID(userID)
		if err != nil {
			return ErrInvalidCredentials
		}
		if !hash.CheckPassword(admin.PasswordHash, req.OldPassword) {
			return ErrInvalidCredentials
		}
		h, err := hash.Password(req.NewPassword)
		if err != nil {
			return err
		}
		admin.PasswordHash = h
		return s.repo.UpdateAdmin(admin)
	}
	student, err := s.repo.FindStudentByID(userID)
	if err != nil {
		return ErrInvalidCredentials
	}
	if !hash.CheckPassword(student.PasswordHash, req.OldPassword) {
		return ErrInvalidCredentials
	}
	h, err := hash.Password(req.NewPassword)
	if err != nil {
		return err
	}
	student.PasswordHash = h
	return s.repo.UpdateStudent(student)
}

// ForgotPassword emails a stateless, signed reset link. Always succeeds (no enumeration).
func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) {
	var userID uuid.UUID
	found := false
	if req.UserType == "admin" {
		if a, err := s.repo.FindAdminByEmail(req.Email); err == nil {
			userID, found = a.ID, true
		}
	} else {
		if st, err := s.repo.FindStudentByEmail(req.Email); err == nil {
			userID, found = st.ID, true
		}
	}
	if !found {
		return
	}
	token, _ := s.signPurposeToken(userID, req.UserType, "reset", 30*time.Minute)
	link := fmt.Sprintf("/reset-password?token=%s", token)
	_ = s.mail.Send(req.Email, "Reset your password",
		"Click the link to reset your password: "+link)
}

// ResetPassword validates a reset token and sets a new password, revoking sessions.
func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	userID, userType, err := s.parsePurposeToken(req.Token, "reset")
	if err != nil {
		return ErrInvalidToken
	}
	h, err := hash.Password(req.NewPassword)
	if err != nil {
		return err
	}
	if userType == "admin" {
		admin, err := s.repo.FindAdminByID(userID)
		if err != nil {
			return ErrInvalidToken
		}
		admin.PasswordHash = h
		if err := s.repo.UpdateAdmin(admin); err != nil {
			return err
		}
	} else {
		student, err := s.repo.FindStudentByID(userID)
		if err != nil {
			return ErrInvalidToken
		}
		student.PasswordHash = h
		if err := s.repo.UpdateStudent(student); err != nil {
			return err
		}
	}
	return s.repo.RevokeAllForUser(userID)
}

// VerifyEmail marks the principal's email as verified.
func (s *AuthService) VerifyEmail(req dto.VerifyEmailRequest) error {
	userID, userType, err := s.parsePurposeToken(req.Token, "verify")
	if err != nil {
		return ErrInvalidToken
	}
	if userType == "admin" {
		a, err := s.repo.FindAdminByID(userID)
		if err != nil {
			return ErrInvalidToken
		}
		a.EmailVerified = true
		return s.repo.UpdateAdmin(a)
	}
	st, err := s.repo.FindStudentByID(userID)
	if err != nil {
		return ErrInvalidToken
	}
	st.EmailVerified = true
	return s.repo.UpdateStudent(st)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *AuthService) issueTokens(claims jwt.Claims, userType jwt.PrincipalType, user dto.AuthUser) (*dto.TokenPair, error) {
	return s.issueTokensWithFamily(claims, userType, user, uuid.New())
}

func (s *AuthService) issueTokensWithFamily(claims jwt.Claims, userType jwt.PrincipalType, user dto.AuthUser, familyID uuid.UUID) (*dto.TokenPair, error) {
	access, err := s.jwtMgr.GenerateAccess(claims)
	if err != nil {
		return nil, err
	}
	raw, err := hash.RandomToken(32)
	if err != nil {
		return nil, err
	}
	rt := &models.RefreshToken{
		UserID:    claims.UserID,
		UserType:  string(userType),
		TokenHash: hash.SHA256(raw),
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshTTL),
	}
	if err := s.repo.CreateRefreshToken(rt); err != nil {
		return nil, err
	}
	return &dto.TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(s.cfg.JWT.AccessTTL.Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) isLocked(until *time.Time) (bool, time.Duration) {
	if until == nil {
		return false, 0
	}
	if time.Now().Before(*until) {
		return true, time.Until(*until)
	}
	return false, 0
}

func (s *AuthService) registerFailure(attempts *int, lockedUntil **time.Time) {
	*attempts++
	if *attempts >= s.cfg.Auth.MaxFailedAttempts {
		t := time.Now().Add(s.cfg.Auth.LockoutDuration)
		*lockedUntil = &t
		*attempts = 0
	}
}

// signPurposeToken builds a short-lived stateless JWT for reset/verify flows.
func (s *AuthService) signPurposeToken(userID uuid.UUID, userType, purpose string, ttl time.Duration) (string, error) {
	claims := gojwt.MapClaims{
		"sub":     userID.String(),
		"type":    userType,
		"purpose": purpose,
		"exp":     time.Now().Add(ttl).Unix(),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(s.cfg.JWT.RefreshSecret))
}

func (s *AuthService) parsePurposeToken(raw, purpose string) (uuid.UUID, string, error) {
	claims := gojwt.MapClaims{}
	tok, err := gojwt.ParseWithClaims(raw, claims, func(t *gojwt.Token) (any, error) {
		return []byte(s.cfg.JWT.RefreshSecret), nil
	})
	if err != nil || !tok.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}
	if claims["purpose"] != purpose {
		return uuid.Nil, "", ErrInvalidToken
	}
	id, err := uuid.Parse(fmt.Sprint(claims["sub"]))
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	return id, fmt.Sprint(claims["type"]), nil
}
