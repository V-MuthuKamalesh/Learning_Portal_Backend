// Package repositories is the data-access layer (GORM). Services depend on these
// interfaces, never on GORM directly.
package repositories

import (
	"errors"
	"time"

	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a lookup yields no row.
var ErrNotFound = errors.New("not found")

// AuthRepository covers identity lookups, lockout bookkeeping and refresh tokens.
type AuthRepository struct{ db *gorm.DB }

func NewAuthRepository(db *gorm.DB) *AuthRepository { return &AuthRepository{db: db} }

func (r *AuthRepository) FindCollegeByCode(code string) (*models.College, error) {
	var c models.College
	err := r.db.Where("code = ? AND is_active = ?", code, true).First(&c).Error
	return wrap(&c, err)
}

// FindStudent locates a student by register number + email, optionally scoped to a college.
func (r *AuthRepository) FindStudent(collegeID *uuid.UUID, register, email string) (*models.Student, error) {
	var s models.Student
	q := r.db.Where("register_number = ? AND email = ?", register, email)
	if collegeID != nil {
		q = q.Where("college_id = ?", *collegeID)
	}
	err := q.First(&s).Error
	return wrap(&s, err)
}

func (r *AuthRepository) FindStudentByID(id uuid.UUID) (*models.Student, error) {
	var s models.Student
	return wrap(&s, r.db.First(&s, "id = ?", id).Error)
}

func (r *AuthRepository) FindStudentByEmail(email string) (*models.Student, error) {
	var s models.Student
	return wrap(&s, r.db.Where("email = ?", email).First(&s).Error)
}

// FindAdmin loads an admin with its role + permissions for JWT claim building.
func (r *AuthRepository) FindAdminByEmail(email string) (*models.Admin, error) {
	var a models.Admin
	err := r.db.Preload("Role.Permissions").Where("email = ?", email).First(&a).Error
	return wrap(&a, err)
}

func (r *AuthRepository) FindAdminByID(id uuid.UUID) (*models.Admin, error) {
	var a models.Admin
	err := r.db.Preload("Role.Permissions").First(&a, "id = ?", id).Error
	return wrap(&a, err)
}

func (r *AuthRepository) UpdateStudent(s *models.Student) error { return r.db.Save(s).Error }
func (r *AuthRepository) UpdateAdmin(a *models.Admin) error     { return r.db.Save(a).Error }

// Refresh token storage.
func (r *AuthRepository) CreateRefreshToken(t *models.RefreshToken) error { return r.db.Create(t).Error }

func (r *AuthRepository) FindRefreshToken(hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	return wrap(&t, r.db.Where("token_hash = ?", hash).First(&t).Error)
}

func (r *AuthRepository) RevokeRefreshToken(id uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

// RevokeFamily revokes every token in a family (reuse detection / logout-all).
func (r *AuthRepository) RevokeFamily(familyID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("family_id = ?", familyID).Update("revoked", true).Error
}

func (r *AuthRepository) RevokeAllForUser(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

// PurgeExpired removes tokens past expiry (call periodically).
func (r *AuthRepository) PurgeExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}

// wrap converts gorm.ErrRecordNotFound into the repo-level ErrNotFound.
func wrap[T any](v *T, err error) (*T, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}
