package services

import (
	"fmt"
	"time"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// AssessmentService implements assessment management.
type AssessmentService struct{ repo *repositories.AssessmentRepository }

func NewAssessmentService(repo *repositories.AssessmentRepository) *AssessmentService {
	return &AssessmentService{repo: repo}
}

func (s *AssessmentService) List(collegeID uuid.UUID) ([]models.Assessment, error) {
	return s.repo.List(collegeID)
}

func (s *AssessmentService) Get(collegeID, id uuid.UUID) (*models.Assessment, error) {
	return s.repo.ByID(collegeID, id)
}

func (s *AssessmentService) Create(collegeID uuid.UUID, adminID uuid.UUID, req dto.CreateAssessmentRequest) (*models.Assessment, error) {
	aType := req.Type
	if aType == "" {
		aType = models.AssessmentMixed
	}
	duration := req.DurationMinutes
	if duration <= 0 {
		duration = 60
	}
	marks := req.TotalMarks
	if marks <= 0 {
		marks = 100
	}
	scoringMode := req.CodingScoringMode
	if scoringMode != "attempt_penalty" {
		scoringMode = "weighted"
	}
	a := &models.Assessment{
		CollegeID:         collegeID,
		Title:             req.Title,
		Description:       req.Description,
		Type:              aType,
		DurationMinutes:   duration,
		TotalMarks:        marks,
		PassingMarks:      req.PassingMarks,
		NegativeMarking:   req.NegativeMarking,
		NegativeMarks:     req.NegativeMarks,
		ShuffleQuestions:  req.ShuffleQuestions,
		CodingScoringMode: scoringMode,
		Status:            models.StatusDraft,
		CreatedBy:         &adminID,
	}
	return a, s.repo.Create(a)
}

func (s *AssessmentService) Update(collegeID, id uuid.UUID, req dto.UpdateAssessmentRequest) (*models.Assessment, error) {
	a, err := s.repo.ByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	applyStr(&a.Title, req.Title)
	applyStr(&a.Description, req.Description)
	applyStr(&a.Type, req.Type)
	applyStr(&a.CodingScoringMode, req.CodingScoringMode)
	if req.DurationMinutes != nil {
		a.DurationMinutes = *req.DurationMinutes
	}
	if req.TotalMarks != nil {
		a.TotalMarks = *req.TotalMarks
	}
	if req.PassingMarks != nil {
		a.PassingMarks = *req.PassingMarks
	}
	if req.NegativeMarking != nil {
		a.NegativeMarking = *req.NegativeMarking
	}
	if req.NegativeMarks != nil {
		a.NegativeMarks = *req.NegativeMarks
	}
	if req.ShuffleQuestions != nil {
		a.ShuffleQuestions = *req.ShuffleQuestions
	}
	if req.StartTime != nil {
		a.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		a.EndTime = req.EndTime
	}
	return a, s.repo.Update(a)
}

func (s *AssessmentService) Delete(collegeID, id uuid.UUID) error {
	return s.repo.Delete(collegeID, id)
}

func (s *AssessmentService) AttachQuestions(collegeID, id uuid.UUID, req dto.AttachQuestionsRequest) error {
	if _, err := s.repo.ByID(collegeID, id); err != nil {
		return err
	}
	ids := make([]uuid.UUID, 0, len(req.QuestionIDs))
	for _, raw := range req.QuestionIDs {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return fmt.Errorf("invalid question id")
		}
		ids = append(ids, parsed)
	}
	return s.repo.AttachQuestions(id, ids)
}

func (s *AssessmentService) Assign(collegeID, id uuid.UUID, req dto.AssignAssessmentRequest) error {
	if _, err := s.repo.ByID(collegeID, id); err != nil {
		return err
	}
	var targetID *uuid.UUID
	if req.TargetID != nil && *req.TargetID != "" {
		parsed, err := uuid.Parse(*req.TargetID)
		if err != nil {
			return fmt.Errorf("invalid target_id")
		}
		targetID = &parsed
	}
	return s.repo.Assign(id, req.TargetType, targetID)
}

func (s *AssessmentService) Publish(collegeID, id uuid.UUID) (*models.Assessment, error) {
	a, err := s.repo.ByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	end := start.Add(7 * 24 * time.Hour)
	if a.StartTime == nil {
		a.StartTime = &start
	}
	if a.EndTime == nil {
		a.EndTime = &end
	}
	if err := s.repo.Publish(id, a.StartTime, a.EndTime); err != nil {
		return nil, err
	}
	a.Status = models.StatusPublished
	return a, nil
}
