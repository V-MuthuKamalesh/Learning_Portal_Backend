package services

import (
	"fmt"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

// QuestionService implements question bank operations.
type QuestionService struct{ repo *repositories.QuestionRepository }

func NewQuestionService(repo *repositories.QuestionRepository) *QuestionService {
	return &QuestionService{repo: repo}
}

func (s *QuestionService) List(collegeID uuid.UUID) ([]models.Question, error) {
	return s.repo.List(collegeID)
}

func (s *QuestionService) CreateMCQ(collegeID uuid.UUID, adminID uuid.UUID, req dto.CreateMCQQuestionRequest) (*models.Question, error) {
	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = models.DifficultyEasy
	}
	marks := req.Marks
	if marks <= 0 {
		marks = 1
	}
	q := &models.Question{
		CollegeID:  collegeID,
		Type:       models.QuestionMCQ,
		Difficulty: difficulty,
		Topic:      req.Topic,
		Marks:      marks,
		CreatedBy:  &adminID,
	}
	mcq := &models.MCQQuestion{
		Body:         req.Body,
		Options:      models.StringSlice(req.Options),
		CorrectIndex: req.CorrectIndex,
		Explanation:  req.Explanation,
	}
	if err := s.repo.Create(q, mcq); err != nil {
		return nil, err
	}
	q.MCQ = mcq
	return q, nil
}

func (s *QuestionService) CreateProgramming(collegeID uuid.UUID, adminID uuid.UUID, req dto.CreateProgrammingQuestionRequest) (*models.Question, error) {
	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = models.DifficultyEasy
	}
	marks := req.Marks
	if marks <= 0 {
		marks = 1
	}
	timeLimit := req.TimeLimitMS
	if timeLimit <= 0 {
		timeLimit = 2000
	}
	memLimit := req.MemoryLimitMB
	if memLimit <= 0 {
		memLimit = 256
	}
	q := &models.Question{
		CollegeID:  collegeID,
		Type:       models.QuestionProgramming,
		Difficulty: difficulty,
		Topic:      req.Topic,
		Marks:      marks,
		CreatedBy:  &adminID,
	}
	prog := &models.ProgrammingQuestion{
		Title:         req.Title,
		Description:   req.Description,
		InputFormat:   req.InputFormat,
		OutputFormat:  req.OutputFormat,
		Constraints:   req.Constraints,
		SampleInput:   req.SampleInput,
		SampleOutput:  req.SampleOutput,
		Explanation:   req.Explanation,
		TimeLimitMS:   timeLimit,
		MemoryLimitMB: memLimit,
	}
	cases := make([]models.TestCase, len(req.TestCases))
	for i, tc := range req.TestCases {
		weight := tc.Weight
		if weight <= 0 {
			weight = 1
		}
		cases[i] = models.TestCase{
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
			IsHidden:       tc.IsHidden,
			Weight:         weight,
		}
	}
	if err := s.repo.CreateProgramming(q, prog, cases); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *QuestionService) Get(collegeID, id uuid.UUID) (*models.Question, error) {
	return s.repo.ByID(collegeID, id)
}

// BulkCreateMCQ imports many MCQ questions, returning counts and per-item errors.
func (s *QuestionService) BulkCreateMCQ(collegeID, adminID uuid.UUID, reqs []dto.CreateMCQQuestionRequest) (created int, errs []string) {
	for i, req := range reqs {
		if _, err := s.CreateMCQ(collegeID, adminID, req); err != nil {
			errs = append(errs, fmt.Sprintf("[%d] %q: %v", i+1, req.Body, err))
		} else {
			created++
		}
	}
	return
}

// BulkCreateProgramming imports many coding questions, returning counts and per-item errors.
func (s *QuestionService) BulkCreateProgramming(collegeID, adminID uuid.UUID, reqs []dto.CreateProgrammingQuestionRequest) (created int, errs []string) {
	for i, req := range reqs {
		if _, err := s.CreateProgramming(collegeID, adminID, req); err != nil {
			errs = append(errs, fmt.Sprintf("[%d] %q: %v", i+1, req.Title, err))
		} else {
			created++
		}
	}
	return
}
