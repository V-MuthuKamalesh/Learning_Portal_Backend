package services

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// GroupService implements group CRUD.
type GroupService struct{ repo *repositories.GroupRepository }

func NewGroupService(repo *repositories.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) List(collegeID uuid.UUID) ([]models.Group, error) {
	return s.repo.List(collegeID)
}

func (s *GroupService) Create(collegeID uuid.UUID, req dto.CreateGroupRequest) (*models.Group, error) {
	g := &models.Group{
		CollegeID:   collegeID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
	}
	if g.Type == "" {
		g.Type = "custom"
	}
	return g, s.repo.Create(g)
}

func (s *GroupService) Update(collegeID, id uuid.UUID, req dto.UpdateGroupRequest) (*models.Group, error) {
	g, err := s.repo.ByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	applyStr(&g.Name, req.Name)
	applyStr(&g.Description, req.Description)
	applyStr(&g.Type, req.Type)
	return g, s.repo.Update(g)
}

func (s *GroupService) Delete(collegeID, id uuid.UUID) error {
	return s.repo.Delete(collegeID, id)
}

func (s *GroupService) ListMembers(collegeID, groupID uuid.UUID) ([]models.Student, error) {
	if _, err := s.repo.ByID(collegeID, groupID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(collegeID, groupID)
}

func (s *GroupService) AddMembers(collegeID, groupID uuid.UUID, studentIDs []uuid.UUID) (int, error) {
	if _, err := s.repo.ByID(collegeID, groupID); err != nil {
		return 0, err
	}
	return s.repo.AddMembers(groupID, studentIDs)
}

func (s *GroupService) RemoveMember(collegeID, groupID, studentID uuid.UUID) error {
	if _, err := s.repo.ByID(collegeID, groupID); err != nil {
		return err
	}
	return s.repo.RemoveMember(groupID, studentID)
}
