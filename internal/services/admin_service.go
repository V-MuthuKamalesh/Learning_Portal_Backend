package services

import (
	"fmt"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/collegeassess/backend/pkg/hash"
	"github.com/google/uuid"
)

// AdminService implements admin management.
type AdminService struct {
	repo     *repositories.AdminRepository
	roleRepo *repositories.RoleRepository
}

func NewAdminService(repo *repositories.AdminRepository, roleRepo *repositories.RoleRepository) *AdminService {
	return &AdminService{repo: repo, roleRepo: roleRepo}
}

func (s *AdminService) List(collegeID uuid.UUID) ([]models.Admin, error) {
	return s.repo.List(collegeID)
}

func (s *AdminService) ListRoles() ([]models.Role, error) {
	return s.roleRepo.ListSystem()
}

func (s *AdminService) Create(collegeID uuid.UUID, req dto.CreateAdminRequest) (*models.Admin, error) {
	exists, err := s.repo.ExistsByEmail(collegeID, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role_id")
	}
	if _, err := s.roleRepo.ByID(roleID); err != nil {
		return nil, fmt.Errorf("role not found")
	}
	pw, err := hash.Password(req.Password)
	if err != nil {
		return nil, err
	}
	a := &models.Admin{
		CollegeID:     collegeID,
		RoleID:        roleID,
		Name:          req.Name,
		Email:         req.Email,
		PasswordHash:  pw,
		Phone:         req.Phone,
		IsActive:      true,
		EmailVerified: true,
	}
	return a, s.repo.Create(a)
}

func (s *AdminService) SetStatus(collegeID, id uuid.UUID, active bool) error {
	a, err := s.repo.ByID(collegeID, id)
	if err != nil {
		return err
	}
	a.IsActive = active
	return s.repo.Update(a)
}

func (s *AdminService) Delete(collegeID, id uuid.UUID) error {
	return s.repo.Delete(collegeID, id)
}
