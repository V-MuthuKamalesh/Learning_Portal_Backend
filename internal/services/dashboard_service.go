package services

import (
	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// DashboardService aggregates tenant metrics.
type DashboardService struct {
	students    *repositories.StudentRepository
	groups      *repositories.GroupRepository
	assessments *repositories.AssessmentRepository
	questions   *repositories.QuestionRepository
}

func NewDashboardService(
	students *repositories.StudentRepository,
	groups *repositories.GroupRepository,
	assessments *repositories.AssessmentRepository,
	questions *repositories.QuestionRepository,
) *DashboardService {
	return &DashboardService{students: students, groups: groups, assessments: assessments, questions: questions}
}

func (s *DashboardService) Stats(collegeID uuid.UUID) (*dto.DashboardStats, error) {
	students, err := s.students.CountByCollege(collegeID)
	if err != nil {
		return nil, err
	}
	groups, err := s.groups.CountByCollege(collegeID)
	if err != nil {
		return nil, err
	}
	assessments, err := s.assessments.CountByCollege(collegeID)
	if err != nil {
		return nil, err
	}
	questions, err := s.questions.CountByCollege(collegeID)
	if err != nil {
		return nil, err
	}
	return &dto.DashboardStats{
		Students:    students,
		Groups:      groups,
		Assessments: assessments,
		Questions:   questions,
	}, nil
}
