package services

import (
	"fmt"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// PracticeService implements practice modules and progress.
type PracticeService struct {
	repo        *repositories.PracticeRepository
	studentRepo *repositories.StudentRepository
	attemptRepo *repositories.AttemptRepository
}

func NewPracticeService(repo *repositories.PracticeRepository, studentRepo *repositories.StudentRepository, attemptRepo *repositories.AttemptRepository) *PracticeService {
	return &PracticeService{repo: repo, studentRepo: studentRepo, attemptRepo: attemptRepo}
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

// AttemptMCQQuestion grades a practice MCQ answer and returns instant feedback.
func (s *PracticeService) AttemptMCQQuestion(collegeID, studentID, moduleID, questionID uuid.UUID, selectedIndex int) (*dto.PracticeAttemptResult, error) {
	mod, err := s.repo.ModuleByID(collegeID, moduleID)
	if err != nil {
		return nil, err
	}
	var found *models.Question
	for i := range mod.Questions {
		if mod.Questions[i].ID == questionID {
			found = &mod.Questions[i]
			break
		}
	}
	if found == nil || found.MCQ == nil {
		return nil, fmt.Errorf("MCQ question not found in module")
	}
	correct := selectedIndex == found.MCQ.CorrectIndex
	marks := 0
	if correct {
		marks = found.Marks
	}
	// Update progress: count this as an attempt toward completion.
	existing, _ := s.repo.ProgressForStudent(studentID)
	completed, total := 0, len(mod.Questions)
	for _, p := range existing {
		if p.ModuleID == moduleID {
			completed = p.Completed
			break
		}
	}
	if correct && completed < total {
		completed++
	}
	_ = s.RecordProgress(studentID, moduleID, completed, total)

	return &dto.PracticeAttemptResult{
		Correct:      correct,
		CorrectIndex: found.MCQ.CorrectIndex,
		Explanation:  found.MCQ.Explanation,
		MarksAwarded: marks,
	}, nil
}

// StudentStats returns aggregated performance stats for the student analytics page.
func (s *PracticeService) StudentStats(studentID uuid.UUID) (*dto.StudentStatsResponse, error) {
	results, err := s.attemptRepo.StudentResults(studentID)
	if err != nil {
		return nil, err
	}
	progress, _ := s.repo.ProgressForStudent(studentID)

	total := len(results)
	passed := 0
	bestScore := 0.0
	sumScore := 0.0
	for _, r := range results {
		sumScore += r.Percentage
		if r.Percentage > bestScore {
			bestScore = r.Percentage
		}
		if r.Passed {
			passed++
		}
	}
	avg := 0.0
	if total > 0 {
		avg = sumScore / float64(total)
	}
	started := len(progress)
	completed := 0
	for _, p := range progress {
		if p.Percentage >= 100 {
			completed++
		}
	}
	return &dto.StudentStatsResponse{
		TotalAttempted:   total,
		AverageScore:     avg,
		BestScore:        bestScore,
		TotalPassed:      passed,
		PracticeModules:  started,
		CompletedModules: completed,
		TopicBreakdown:   []dto.TopicStat{},
	}, nil
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
