package services

import (
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// PracticeService implements practice modules and progress.
type PracticeService struct {
	repo        *repositories.PracticeRepository
	studentRepo *repositories.StudentRepository
}

func NewPracticeService(repo *repositories.PracticeRepository, studentRepo *repositories.StudentRepository) *PracticeService {
	return &PracticeService{repo: repo, studentRepo: studentRepo}
}

func (s *PracticeService) ListModules(collegeID uuid.UUID) ([]models.PracticeModule, error) {
	return s.repo.ListModules(collegeID)
}

func (s *PracticeService) ModuleDetail(collegeID, moduleID uuid.UUID) (*models.PracticeModule, error) {
	return s.repo.ModuleByID(collegeID, moduleID)
}

func (s *PracticeService) StudentProgress(studentID uuid.UUID) ([]models.StudentProgress, error) {
	return s.repo.ProgressForStudent(studentID)
}

func (s *PracticeService) RecordProgress(studentID, moduleID uuid.UUID, completed, total int) error {
	pct := 0.0
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}
	return s.repo.UpsertProgress(&models.StudentProgress{
		StudentID:  studentID,
		ModuleID:   moduleID,
		Completed:  completed,
		Total:      total,
		Percentage: pct,
	})
}

// NotificationService lists and marks notifications.
type NotificationService struct{ repo *repositories.NotificationRepository }

func NewNotificationService(repo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) List(userID uuid.UUID, userType string) ([]models.Notification, error) {
	return s.repo.ListForUser(userID, userType)
}

func (s *NotificationService) MarkRead(id, userID uuid.UUID) error {
	return s.repo.MarkRead(id, userID)
}
