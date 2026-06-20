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
	_, prog, marks, _, err := s.loadProgrammingContext(submissionID, studentID, assessmentQuestionID)
	if err != nil {
		return nil, err
	}
	// Run-mode: public tests only, no penalty, not graded
	return s.evaluateCode(prog, marks, req.Language, req.SourceCode, false, "weighted", 0)
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
	aq, prog, marks, scoringMode, err := s.loadProgrammingContext(submissionID, studentID, assessmentQuestionID)
	if err != nil {
		return nil, err
	}

	// Determine prior failed attempts for penalty calculation.
	priorFailed := 0
	priorAttemptCount := 0
	existing, lookupErr := s.attemptRepo.FindCodingSubmission(submissionID, assessmentQuestionID)
	if lookupErr == nil {
		priorAttemptCount = existing.AttemptCount
		priorFailed = existing.FailedAttempts
		// If the last attempt also wasn't fully accepted, count it as a failure.
		if existing.Status != models.JudgeAccepted {
			priorFailed = existing.FailedAttempts + 1
		} else {
			// Already accepted — keep previous failed count (don't add more).
			priorFailed = existing.FailedAttempts
		}
	}

	result, err := s.evaluateCode(prog, marks, req.Language, req.SourceCode, true, scoringMode, priorFailed)
	if err != nil {
		return nil, err
	}

	// Compute new failure count: if this attempt is not fully accepted, it's a failure.
	newFailed := priorFailed
	if lookupErr != nil {
		// First attempt — no prior failures contributed yet.
		newFailed = 0
		if result.Status != models.JudgeAccepted {
			newFailed = 1
		}
	}

	newAttemptCount := priorAttemptCount + 1
	if newAttemptCount <= 0 {
		newAttemptCount = 1
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
		AttemptCount:         newAttemptCount,
		FailedAttempts:       newFailed,
	}
	if err := s.attemptRepo.UpsertCodingSubmission(cs); err != nil {
		return nil, err
	}
	result.AttemptCount = newAttemptCount
	result.FailedAttempts = newFailed
	return result, nil
}

// loadProgrammingContext validates the attempt and loads the programming question + scoring config.
func (s *CodingService) loadProgrammingContext(
	submissionID, studentID, assessmentQuestionID uuid.UUID,
) (*models.AssessmentQuestion, *models.ProgrammingQuestion, int, string, error) {
	sub, err := s.attemptRepo.SubmissionByID(submissionID)
	if err != nil {
		return nil, nil, 0, "", err
	}
	if sub.StudentID != studentID || sub.Status != models.SubInProgress {
		return nil, nil, 0, "", ErrAttemptClosed
	}
	st, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, nil, 0, "", err
	}
	a, err := s.assessRepo.ByID(st.CollegeID, sub.AssessmentID)
	if err != nil {
		return nil, nil, 0, "", err
	}
	var aq *models.AssessmentQuestion
	for i := range a.Questions {
		if a.Questions[i].ID == assessmentQuestionID {
			aq = &a.Questions[i]
			break
		}
	}
	if aq == nil || aq.Question == nil || aq.Question.Programming == nil {
		return nil, nil, 0, "", fmt.Errorf("programming question not found")
	}
	marks := a.TotalMarks / max(len(a.Questions), 1)
	if aq.Marks != nil {
		marks = *aq.Marks
	}
	scoringMode := a.CodingScoringMode
	if scoringMode == "" {
		scoringMode = "weighted"
	}
	return aq, aq.Question.Programming, marks, scoringMode, nil
}

// evaluateCode runs the source against test cases and computes marks.
// scoringMode: "weighted" | "attempt_penalty"
// failedAttempts: number of prior failed submissions (for attempt_penalty mode)
func (s *CodingService) evaluateCode(
	prog *models.ProgrammingQuestion,
	totalMarks int,
	language, source string,
	includeHidden bool,
	scoringMode string,
	failedAttempts int,
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
		switch out.Status {
		case "tle":
			row.Status = models.JudgeTLE
		case "mle":
			row.Status = models.JudgeMLE
		case "error":
			row.Status = models.JudgeError
		default:
			if judgepkg.OutputsMatch(out.Stdout, tc.ExpectedOutput) {
				row.Passed = true
				row.Status = models.JudgeAccepted
				passed++
				earnedWeight += weight
			} else {
				row.Status = models.JudgeWrong
			}
		}
		results = append(results, row)
	}

	status := models.JudgeWrong
	if passed == total && total > 0 {
		status = models.JudgeAccepted
	} else if passed > 0 {
		status = models.JudgePartial
	}

	// Calculate marks based on scoring mode.
	passRatio := float64(earnedWeight) / float64(totalWeight)
	var marksAwarded float64
	switch scoringMode {
	case "attempt_penalty":
		// Reduce max marks by 10% per prior failed attempt (floor 0).
		penalty := float64(failedAttempts) * 0.1
		if penalty > 1 {
			penalty = 1
		}
		effectiveMax := float64(totalMarks) * (1 - penalty)
		marksAwarded = effectiveMax * passRatio
	default: // "weighted"
		marksAwarded = float64(totalMarks) * passRatio
	}
	if marksAwarded < 0 {
		marksAwarded = 0
	}

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
