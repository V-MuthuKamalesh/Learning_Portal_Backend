package repositories

import (
	"github.com/collegeassess/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminRepository handles admin persistence.
type AdminRepository struct{ db *gorm.DB }

func NewAdminRepository(db *gorm.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) Create(a *models.Admin) error { return r.db.Create(a).Error }
func (r *AdminRepository) Update(a *models.Admin) error { return r.db.Save(a).Error }

func (r *AdminRepository) List(collegeID uuid.UUID) ([]models.Admin, error) {
	var admins []models.Admin
	err := r.db.Preload("Role").Where("college_id = ?", collegeID).Order("created_at DESC").Find(&admins).Error
	return admins, err
}

func (r *AdminRepository) ByID(collegeID, id uuid.UUID) (*models.Admin, error) {
	var a models.Admin
	err := r.db.Preload("Role").Where("college_id = ? AND id = ?", collegeID, id).First(&a).Error
	return wrap(&a, err)
}

func (r *AdminRepository) ExistsByEmail(collegeID uuid.UUID, email string) (bool, error) {
	var n int64
	err := r.db.Model(&models.Admin{}).Where("college_id = ? AND email = ?", collegeID, email).Count(&n).Error
	return n > 0, err
}

func (r *AdminRepository) Delete(collegeID, id uuid.UUID) error {
	return r.db.Where("college_id = ?", collegeID).Delete(&models.Admin{}, "id = ?", id).Error
}

// RoleRepository lists assignable roles.
type RoleRepository struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) *RoleRepository { return &RoleRepository{db: db} }

func (r *RoleRepository) ListSystem() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Where("is_system = ?", true).Order("name").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) ByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	return wrap(&role, r.db.First(&role, "id = ?", id).Error)
}
