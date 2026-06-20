package services

import (
	"fmt"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// AdminPracticeService manages practice modules from the admin side.
type AdminPracticeService struct {
	practice  *repositories.PracticeRepository
	questions *repositories.QuestionRepository
}

func NewAdminPracticeService(pr *repositories.PracticeRepository, qr *repositories.QuestionRepository) *AdminPracticeService {
	return &AdminPracticeService{practice: pr, questions: qr}
}

func (s *AdminPracticeService) CreateModule(collegeID uuid.UUID, req dto.CreatePracticeModuleRequest) (*models.PracticeModule, error) {
	cat := req.Category
	if cat == "" {
		cat = "mixed"
	}
	m := &models.PracticeModule{
		CollegeID:   collegeID,
		Name:        req.Name,
		Category:    cat,
		Description: req.Description,
		Tags:        req.Tags,
		Ord:         req.Ord,
		IsPublished: false,
	}
	if err := s.practice.CreateModule(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *AdminPracticeService) ListModules(collegeID uuid.UUID) ([]models.PracticeModule, error) {
	return s.practice.ListModules(collegeID)
}

func (s *AdminPracticeService) GetModule(collegeID, id uuid.UUID) (*models.PracticeModule, error) {
	return s.practice.ModuleByID(collegeID, id)
}

func (s *AdminPracticeService) UpdateModule(collegeID, id uuid.UUID, req dto.UpdatePracticeModuleRequest) (*models.PracticeModule, error) {
	m, err := s.practice.ModuleByID(collegeID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Category != "" {
		m.Category = req.Category
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.Tags != "" {
		m.Tags = req.Tags
	}
	if req.IsPublished != nil {
		m.IsPublished = *req.IsPublished
	}
	if req.Ord != 0 {
		m.Ord = req.Ord
	}
	if err := s.practice.UpdateModule(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *AdminPracticeService) DeleteModule(collegeID, id uuid.UUID) error {
	return s.practice.DeleteModule(collegeID, id)
}

// ListModuleQuestions returns all question slots with enriched question data.
func (s *AdminPracticeService) ListModuleQuestions(collegeID, moduleID uuid.UUID) ([]dto.ModuleQuestionDetail, error) {
	slots, qs, err := s.practice.ListModuleQuestions(collegeID, moduleID)
	if err != nil {
		return nil, err
	}
	qMap := map[uuid.UUID]*models.Question{}
	for i := range qs {
		qMap[qs[i].ID] = &qs[i]
	}

	out := make([]dto.ModuleQuestionDetail, len(slots))
	for i, sl := range slots {
		d := dto.ModuleQuestionDetail{
			ID:          sl.ID.String(),
			QuestionID:  sl.QuestionID.String(),
			Ord:         sl.Ord,
			Marks:       sl.Marks,
			MaxAttempts: sl.MaxAttempts,
		}
		if q, ok := qMap[sl.QuestionID]; ok {
			summary := &dto.QuestionSummary{
				ID:         q.ID.String(),
				Type:       q.Type,
				Topic:      q.Topic,
				Difficulty: q.Difficulty,
				Marks:      q.Marks,
			}
			if q.MCQ != nil {
				summary.Body = q.MCQ.Body
			}
			if q.Programming != nil {
				summary.Title = q.Programming.Title
			}
			d.Question = summary
		}
		out[i] = d
	}
	return out, nil
}

// AddQuestions upserts question slots into a module (idempotent).
func (s *AdminPracticeService) AddQuestions(collegeID, moduleID uuid.UUID, req dto.AddModuleQuestionsRequest) (added int, errs []string) {
	// verify module belongs to college
	if _, err := s.practice.ModuleByID(collegeID, moduleID); err != nil {
		return 0, []string{"module not found"}
	}
	for i, entry := range req.Questions {
		qid, err := uuid.Parse(entry.QuestionID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%d] invalid question_id: %s", i+1, entry.QuestionID))
			continue
		}
		// verify question belongs to college
		if _, err := s.questions.ByID(collegeID, qid); err != nil {
			errs = append(errs, fmt.Sprintf("[%d] question %s not found", i+1, entry.QuestionID))
			continue
		}
		if err := s.practice.LinkQuestion(moduleID, qid, entry.Ord, entry.Marks, entry.MaxAttempts); err != nil {
			errs = append(errs, fmt.Sprintf("[%d] failed to add %s: %v", i+1, entry.QuestionID, err))
			continue
		}
		added++
	}
	return
}

// RemoveQuestion unlinks a question from a module.
func (s *AdminPracticeService) RemoveQuestion(collegeID, moduleID uuid.UUID, questionIDStr string) error {
	if _, err := s.practice.ModuleByID(collegeID, moduleID); err != nil {
		return err
	}
	qid, err := uuid.Parse(questionIDStr)
	if err != nil {
		return fmt.Errorf("invalid question_id")
	}
	return s.practice.UnlinkQuestion(moduleID, qid)
}

// UpdateQuestionSlot patches marks / max_attempts / ord for one slot.
func (s *AdminPracticeService) UpdateQuestionSlot(collegeID, moduleID uuid.UUID, questionIDStr string, req dto.UpdateModuleQuestionRequest) error {
	if _, err := s.practice.ModuleByID(collegeID, moduleID); err != nil {
		return err
	}
	qid, err := uuid.Parse(questionIDStr)
	if err != nil {
		return fmt.Errorf("invalid question_id")
	}
	slot, err := s.practice.ModuleQuestionSlot(moduleID, qid)
	if err != nil {
		return err
	}
	slot.Marks = req.Marks
	slot.MaxAttempts = req.MaxAttempts
	if req.Ord != 0 {
		slot.Ord = req.Ord
	}
	return s.practice.UpdateModuleQuestion(slot)
}

// ReorderQuestions sets question ordering by providing question IDs in desired order.
func (s *AdminPracticeService) ReorderQuestions(collegeID, moduleID uuid.UUID, req dto.ReorderModuleQuestionsRequest) error {
	if _, err := s.practice.ModuleByID(collegeID, moduleID); err != nil {
		return err
	}
	for i, qidStr := range req.QuestionIDs {
		qid, err := uuid.Parse(qidStr)
		if err != nil {
			continue
		}
		slot, err := s.practice.ModuleQuestionSlot(moduleID, qid)
		if err != nil {
			continue
		}
		slot.Ord = i
		_ = s.practice.UpdateModuleQuestion(slot)
	}
	return nil
}
