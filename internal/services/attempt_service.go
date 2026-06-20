package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/collegeassess/backend/internal/dto"
	"github.com/collegeassess/backend/internal/models"
	"github.com/collegeassess/backend/internal/repositories"
	"github.com/google/uuid"
)

var (
	ErrAttemptClosed   = errors.New("attempt is closed")
	ErrAlreadySubmitted = errors.New("attempt already submitted")
)

// AttemptService handles student assessment attempts and grading.
type AttemptService struct {
	assessRepo *repositories.AssessmentRepository
	attemptRepo *repositories.AttemptRepository
	studentRepo *repositories.StudentRepository
	notifRepo   *repositories.NotificationRepository
}

func NewAttemptService(
	assessRepo *repositories.AssessmentRepository,
	attemptRepo *repositories.AttemptRepository,
	studentRepo *repositories.StudentRepository,
	notifRepo *repositories.NotificationRepository,
) *AttemptService {
	return &AttemptService{
		assessRepo: assessRepo, attemptRepo: attemptRepo,
		studentRepo: studentRepo, notifRepo: notifRepo,
	}
}

func (s *AttemptService) ListAssessments(collegeID, studentID uuid.UUID) ([]dto.StudentAssessmentView, error) {
	st, err := s.studentRepo.ByID(collegeID, studentID)
	if err != nil {
		return nil, err
	}
	groupIDs := make([]uuid.UUID, 0, len(st.Groups))
	for _, g := range st.Groups {
		groupIDs = append(groupIDs, g.ID)
	}
	items, err := s.assessRepo.ListForStudent(collegeID, studentID, st.DepartmentID, st.BatchID, groupIDs)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	out := make([]dto.StudentAssessmentView, 0, len(items))
	for _, a := range items {
		view := dto.StudentAssessmentView{
			ID: a.ID.String(), Title: a.Title, Description: a.Description,
			Type: a.Type, Status: a.Status, Workflow: a.Workflow(now),
			DurationMinutes: a.DurationMinutes, TotalMarks: a.TotalMarks,
			StartTime: a.StartTime, EndTime: a.EndTime,
		}
		if sub, err := s.attemptRepo.FindSubmission(a.ID, studentID); err == nil {
			sid := sub.ID.String()
			view.SubmissionID = &sid
			view.AttemptStatus = sub.Status
			if sub.Status == models.SubGraded || sub.Status == models.SubSubmitted {
				score := float64(sub.TotalScore)
				view.Score = &score
			}
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *AttemptService) Start(collegeID, studentID, assessmentID uuid.UUID) (*dto.AttemptDetail, error) {
	a, err := s.assessRepo.ByID(collegeID, assessmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != models.StatusPublished {
		return nil, fmt.Errorf("assessment not published")
	}
	sub, err := s.attemptRepo.FindSubmission(assessmentID, studentID)
	if err != nil {
		if !errors.Is(err, repositories.ErrNotFound) {
			return nil, err
		}
		sub = &models.Submission{
			AssessmentID: assessmentID,
			StudentID:    studentID,
			Status:       models.SubInProgress,
			StartedAt:    time.Now(),
		}
		if err := s.attemptRepo.CreateSubmission(sub); err != nil {
			return nil, err
		}
	}
	if sub.Status != models.SubInProgress {
		return nil, ErrAlreadySubmitted
	}
	return s.buildAttemptDetail(a, sub), nil
}

func (s *AttemptService) GetAttempt(submissionID, studentID uuid.UUID) (*dto.AttemptDetail, error) {
	sub, err := s.attemptRepo.SubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}
	if sub.StudentID != studentID {
		return nil, repositories.ErrNotFound
	}
	st, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, err
	}
	a, err := s.assessRepo.ByID(st.CollegeID, sub.AssessmentID)
	if err != nil {
		return nil, err
	}
	return s.buildAttemptDetail(a, sub), nil
}

func (s *AttemptService) buildAttemptDetail(a *models.Assessment, sub *models.Submission) *dto.AttemptDetail {
	expires := sub.StartedAt.Add(time.Duration(a.DurationMinutes) * time.Minute)
	answers := map[string]int{}
	for _, ans := range sub.Answers {
		if ans.SelectedIndex != nil {
			answers[ans.AssessmentQuestionID.String()] = *ans.SelectedIndex
		}
	}
	coding := map[string]dto.CodingSubmissionView{}
	for _, cs := range sub.CodingSubmissions {
		key := ""
		if cs.AssessmentQuestionID != nil {
			key = cs.AssessmentQuestionID.String()
		}
		if key == "" {
			continue
		}
		coding[key] = dto.CodingSubmissionView{
			Language:       cs.Language,
			SourceCode:     cs.SourceCode,
			Status:         cs.Status,
			PassedCases:    cs.PassedCases,
			TotalCases:     cs.TotalCases,
			MarksAwarded:   cs.MarksAwarded,
			AttemptCount:   cs.AttemptCount,
			FailedAttempts: cs.FailedAttempts,
		}
	}
	questions := make([]dto.AttemptQuestion, 0, len(a.Questions))
	for _, aq := range a.Questions {
		if aq.Question == nil {
			continue
		}
		marks := a.TotalMarks / max(len(a.Questions), 1)
		if aq.Marks != nil {
			marks = *aq.Marks
		}
		if aq.Question.Type == models.QuestionMCQ && aq.Question.MCQ != nil {
			questions = append(questions, dto.AttemptQuestion{
				ID:                   aq.QuestionID.String(),
				AssessmentQuestionID: aq.ID.String(),
				Type:                 models.QuestionMCQ,
				Body:                 aq.Question.MCQ.Body,
				Options:              []string(aq.Question.MCQ.Options),
				Marks:                marks,
				Ord:                  aq.Ord,
			})
			continue
		}
		if aq.Question.Type == models.QuestionProgramming && aq.Question.Programming != nil {
			p := aq.Question.Programming
			questions = append(questions, dto.AttemptQuestion{
				ID:                   aq.QuestionID.String(),
				AssessmentQuestionID: aq.ID.String(),
				Type:                 models.QuestionProgramming,
				Title:                p.Title,
				Description:          p.Description,
				InputFormat:          p.InputFormat,
				OutputFormat:         p.OutputFormat,
				Constraints:          p.Constraints,
				SampleInput:          p.SampleInput,
				SampleOutput:         p.SampleOutput,
				TimeLimitMS:          p.TimeLimitMS,
				MemoryLimitMB:        p.MemoryLimitMB,
				Marks:                marks,
				Ord:                  aq.Ord,
			})
		}
	}
	scoringMode := a.CodingScoringMode
	if scoringMode == "" {
		scoringMode = "weighted"
	}
	return &dto.AttemptDetail{
		ID: sub.ID.String(), AssessmentID: a.ID.String(), Status: sub.Status,
		StartedAt: sub.StartedAt, ExpiresAt: expires,
		Questions: questions, Answers: answers, Coding: coding,
		CodingScoringMode: scoringMode,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *AttemptService) SaveAnswer(submissionID, studentID uuid.UUID, req dto.SaveAnswerRequest) error {
	sub, err := s.attemptRepo.SubmissionByID(submissionID)
	if err != nil {
		return err
	}
	if sub.StudentID != studentID || sub.Status != models.SubInProgress {
		return ErrAttemptClosed
	}
	aqID, err := uuid.Parse(req.AssessmentQuestionID)
	if err != nil {
		return fmt.Errorf("invalid assessment_question_id")
	}
	idx := req.SelectedIndex
	return s.attemptRepo.UpsertAnswer(&models.Answer{
		SubmissionID:         submissionID,
		AssessmentQuestionID: aqID,
		SelectedIndex:        &idx,
	})
}

func (s *AttemptService) Submit(submissionID, studentID uuid.UUID) (*models.AssessmentResult, error) {
	sub, err := s.attemptRepo.SubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}
	if sub.StudentID != studentID {
		return nil, repositories.ErrNotFound
	}
	if sub.Status != models.SubInProgress {
		return nil, ErrAlreadySubmitted
	}
	st, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, err
	}
	a, err := s.assessRepo.ByID(st.CollegeID, sub.AssessmentID)
	if err != nil {
		return nil, err
	}

	answerMap := map[uuid.UUID]int{}
	for _, ans := range sub.Answers {
		if ans.SelectedIndex != nil {
			answerMap[ans.AssessmentQuestionID] = *ans.SelectedIndex
		}
	}

	var scored float64
	correct, wrong := 0, 0
	perQ := float64(a.TotalMarks) / float64(max(len(a.Questions), 1))

	codingByAQ := map[uuid.UUID]models.CodingSubmission{}
	for _, cs := range sub.CodingSubmissions {
		if cs.AssessmentQuestionID != nil {
			codingByAQ[*cs.AssessmentQuestionID] = cs
		}
	}

	for _, aq := range a.Questions {
		if aq.Question == nil {
			continue
		}
		marks := perQ
		if aq.Marks != nil {
			marks = float64(*aq.Marks)
		}
		if aq.Question.Type == models.QuestionMCQ && aq.Question.MCQ != nil {
			selected, ok := answerMap[aq.ID]
			isCorrect := ok && selected == aq.Question.MCQ.CorrectIndex
			awarded := 0.0
			if isCorrect {
				awarded = marks
				correct++
			} else if ok {
				wrong++
				if a.NegativeMarking {
					awarded = -a.NegativeMarks
				}
			}
			scored += awarded
			if ansIdx, exists := answerMap[aq.ID]; exists {
				correctVal := isCorrect
				_ = s.attemptRepo.UpsertAnswer(&models.Answer{
					SubmissionID:         submissionID,
					AssessmentQuestionID: aq.ID,
					SelectedIndex:        &ansIdx,
					IsCorrect:            &correctVal,
					MarksAwarded:         awarded,
				})
			}
			continue
		}
		if aq.Question.Type == models.QuestionProgramming {
			if cs, ok := codingByAQ[aq.ID]; ok {
				scored += cs.MarksAwarded
				if cs.Status == models.JudgeAccepted {
					correct++
				} else if cs.PassedCases > 0 {
					correct++
				} else {
					wrong++
				}
			} else {
				wrong++
			}
		}
	}

	now := time.Now()
	sub.Status = models.SubSubmitted
	sub.SubmittedAt = &now
	sub.TotalScore = int(scored)
	if err := s.attemptRepo.UpdateSubmission(sub); err != nil {
		return nil, err
	}

	pct := 0.0
	if a.TotalMarks > 0 {
		pct = (scored / float64(a.TotalMarks)) * 100
	}
	result := &models.AssessmentResult{
		AssessmentID: sub.AssessmentID,
		StudentID:    studentID,
		MarksScored:  scored,
		TotalMarks:   a.TotalMarks,
		Percentage:   pct,
		CorrectCount: correct,
		WrongCount:   wrong,
		Passed:       scored >= float64(a.PassingMarks),
		Published:    true,
	}
	if err := s.attemptRepo.CreateResult(result); err != nil {
		return nil, err
	}
	_ = s.attemptRepo.RecalculateRanks(sub.AssessmentID)

	sub.Status = models.SubGraded
	_ = s.attemptRepo.UpdateSubmission(sub)

	_ = s.notifRepo.Create(&models.Notification{
		CollegeID: st.CollegeID,
		UserID:    studentID,
		UserType:  "student",
		Type:      models.NotifResultPublished,
		Title:     "Result available",
		Body:      fmt.Sprintf("Your result for %s is ready (%.0f%%).", a.Title, pct),
		Link:      "/results",
	})

	return result, nil
}

func (s *AttemptService) StudentResults(studentID uuid.UUID) ([]models.AssessmentResult, error) {
	return s.attemptRepo.StudentResults(studentID)
}

func (s *AttemptService) Leaderboard(collegeID uuid.UUID) ([]models.AssessmentResult, error) {
	return s.attemptRepo.Leaderboard(collegeID, 20)
}

func (s *AttemptService) AdminResults(collegeID uuid.UUID, assessmentID *uuid.UUID) ([]models.AssessmentResult, error) {
	return s.attemptRepo.ListResults(collegeID, assessmentID)
}
