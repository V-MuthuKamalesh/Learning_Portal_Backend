package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	judgepkg "github.com/collegeassess/backend/pkg/judge"
	"github.com/google/uuid"
)

// CodingService runs and grades programming submissions via the judge.
type CodingService struct {
	judge        *judgepkg.Client
	attemptRepo  *repositories.AttemptRepository
	assessRepo   *repositories.AssessmentRepository
	studentRepo  *repositories.StudentRepository
}

func NewCodingService(
	judge *judgepkg.Client,
	attemptRepo *repositories.AttemptRepository,
	assessRepo *repositories.AssessmentRepository,
	studentRepo *repositories.StudentRepository,
) *CodingService {
	return &CodingService{
		judge: judge, attemptRepo: attemptRepo,
		assessRepo: assessRepo, studentRepo: studentRepo,
	}
}

func (s *CodingService) RunCode(req dto.RunCodeRequest) (*dto.CodingRunResult, error) {
	if !s.judge.Enabled() {
		return nil, errors.New("judge service is disabled")
	}
	timeLimit := req.TimeLimitMS
	if timeLimit <= 0 {
		timeLimit = 2000
	}
	out, err := s.judge.Execute(judgepkg.ExecuteRequest{
		Language:    req.Language,
		Source:      req.Source,
		Stdin:       req.Stdin,
		TimeLimitMS: timeLimit,
	})
	if err != nil {
		return nil, err
	}
	status := models.JudgeAccepted
	if out.Status == "tle" {
		status = models.JudgeTLE
	} else if out.Status != "ok" {
		status = models.JudgeError
	}
	return &dto.CodingRunResult{
		Status:    status,
		Stdout:    out.Stdout,
		Stderr:    out.Stderr,
		RuntimeMS: out.RuntimeMS,
	}, nil
}

func (s *CodingService) RunAttemptQuestion(
	submissionID, studentID, assessmentQuestionID uuid.UUID,
	req dto.SubmitCodingRequest,
) (*dto.CodingRunResult, error) {
	_, prog, marks, err := s.loadProgrammingContext(submissionID, studentID, assessmentQuestionID)
	if err != nil {
		return nil, err
	}
	return s.evaluateCode(prog, marks, req.Language, req.SourceCode, false)
}

func (s *CodingService) SubmitAttemptQuestion(
	submissionID, studentID, assessmentQuestionID uuid.UUID,
	req dto.SubmitCodingRequest,
) (*dto.CodingRunResult, error) {
	sub, err := s.attemptRepo.SubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}
	if sub.StudentID != studentID || sub.Status != models.SubInProgress {
		return nil, ErrAttemptClosed
	}
	aq, prog, marks, err := s.loadProgrammingContext(submissionID, studentID, assessmentQuestionID)
	if err != nil {
		return nil, err
	}
	result, err := s.evaluateCode(prog, marks, req.Language, req.SourceCode, true)
	if err != nil {
		return nil, err
	}

	verdictJSON, _ := json.Marshal(map[string]any{"results": result.Results})
	var verdict models.JSON
	_ = json.Unmarshal(verdictJSON, &verdict)
	cs := &models.CodingSubmission{
		SubmissionID:         &submissionID,
		AssessmentQuestionID: &aq.ID,
		QuestionID:           aq.QuestionID,
		StudentID:            studentID,
		Language:             req.Language,
		SourceCode:           req.SourceCode,
		Status:               result.Status,
		PassedCases:          result.PassedCases,
		TotalCases:           result.TotalCases,
		MarksAwarded:         result.MarksAwarded,
		RuntimeMS:            result.RuntimeMS,
		Verdict:              verdict,
	}
	if err := s.attemptRepo.UpsertCodingSubmission(cs); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *CodingService) loadProgrammingContext(
	submissionID, studentID, assessmentQuestionID uuid.UUID,
) (*models.AssessmentQuestion, *models.ProgrammingQuestion, int, error) {
	sub, err := s.attemptRepo.SubmissionByID(submissionID)
	if err != nil {
		return nil, nil, 0, err
	}
	if sub.StudentID != studentID || sub.Status != models.SubInProgress {
		return nil, nil, 0, ErrAttemptClosed
	}
	st, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, nil, 0, err
	}
	a, err := s.assessRepo.ByID(st.CollegeID, sub.AssessmentID)
	if err != nil {
		return nil, nil, 0, err
	}
	var aq *models.AssessmentQuestion
	for i := range a.Questions {
		if a.Questions[i].ID == assessmentQuestionID {
			aq = &a.Questions[i]
			break
		}
	}
	if aq == nil || aq.Question == nil || aq.Question.Programming == nil {
		return nil, nil, 0, fmt.Errorf("programming question not found")
	}
	marks := a.TotalMarks / max(len(a.Questions), 1)
	if aq.Marks != nil {
		marks = *aq.Marks
	}
	return aq, aq.Question.Programming, marks, nil
}

func (s *CodingService) evaluateCode(
	prog *models.ProgrammingQuestion,
	totalMarks int,
	language, source string,
	includeHidden bool,
) (*dto.CodingRunResult, error) {
	if !s.judge.Enabled() {
		return nil, errors.New("judge service is disabled")
	}
	cases := prog.TestCases
	if len(cases) == 0 {
		return nil, fmt.Errorf("no test cases configured")
	}

	totalWeight := 0
	for _, tc := range cases {
		if includeHidden || !tc.IsHidden {
			w := tc.Weight
			if w <= 0 {
				w = 1
			}
			totalWeight += w
		}
	}
	if totalWeight == 0 {
		totalWeight = 1
	}

	results := make([]dto.TestCaseResult, 0, len(cases))
	passed := 0
	total := 0
	var earnedWeight int
	var lastStdout, lastStderr string
	var lastRuntime int

	for _, tc := range cases {
		if !includeHidden && tc.IsHidden {
			continue
		}
		total++
		weight := tc.Weight
		if weight <= 0 {
			weight = 1
		}
		out, err := s.judge.Execute(judgepkg.ExecuteRequest{
			Language:      language,
			Source:        source,
			Stdin:         tc.Input,
			TimeLimitMS:   prog.TimeLimitMS,
			MemoryLimitMB: prog.MemoryLimitMB,
		})
		row := dto.TestCaseResult{
			Ord:            tc.Ord,
			IsHidden:       tc.IsHidden,
			Weight:         weight,
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
		}
		if err != nil {
			row.Status = models.JudgeError
			row.ActualOutput = err.Error()
			results = append(results, row)
			continue
		}
		lastStdout = out.Stdout
		lastStderr = out.Stderr
		lastRuntime = out.RuntimeMS
		row.ActualOutput = out.Stdout
		if out.Status == "tle" {
			row.Status = models.JudgeTLE
		} else if out.Status == "error" {
			row.Status = models.JudgeError
		} else if judgepkg.OutputsMatch(out.Stdout, tc.ExpectedOutput) {
			row.Passed = true
			row.Status = models.JudgeAccepted
			passed++
			earnedWeight += weight
		} else {
			row.Status = models.JudgeWrong
		}
		results = append(results, row)
	}

	status := models.JudgeWrong
	if passed == total && total > 0 {
		status = models.JudgeAccepted
	} else if passed > 0 {
		status = models.JudgePartial
	}
	marksAwarded := float64(totalMarks) * float64(earnedWeight) / float64(totalWeight)

	return &dto.CodingRunResult{
		Status:       status,
		PassedCases:  passed,
		TotalCases:   total,
		MarksAwarded: marksAwarded,
		Results:      results,
		Stdout:       lastStdout,
		Stderr:       lastStderr,
		RuntimeMS:    lastRuntime,
	}, nil
}
