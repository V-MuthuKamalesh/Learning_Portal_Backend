package models

import (
	"time"

	"github.com/google/uuid"
)

// Submission/grading status enums.
const (
	SubInProgress    = "in_progress"
	SubSubmitted     = "submitted"
	SubAutoSubmitted = "auto_submitted"
	SubGraded        = "graded"

	JudgeAccepted = "accepted"
	JudgePartial  = "partial"
	JudgeWrong    = "wrong"
	JudgeError    = "error"
	JudgeTLE      = "tle"
	JudgeMLE      = "mle"
	JudgePending  = "pending"
)

// Submission is one student's attempt at an assessment.
type Submission struct {
	Base
	AssessmentID uuid.UUID  `gorm:"type:uuid;not null;index:idx_sub_unique,unique" json:"assessment_id"`
	StudentID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_sub_unique,unique" json:"student_id"`
	Status       string     `gorm:"size:20;not null;default:'in_progress'" json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	TotalScore   int        `gorm:"default:0" json:"total_score"`

	Answers          []Answer           `gorm:"foreignKey:SubmissionID" json:"answers,omitempty"`
	CodingSubmissions []CodingSubmission `gorm:"foreignKey:SubmissionID" json:"coding_submissions,omitempty"`
}

// Answer is a student's MCQ selection within a submission.
type Answer struct {
	Base
	SubmissionID         uuid.UUID `gorm:"type:uuid;not null;index:idx_ans_unique,unique" json:"submission_id"`
	AssessmentQuestionID uuid.UUID `gorm:"type:uuid;not null;index:idx_ans_unique,unique" json:"assessment_question_id"`
	SelectedIndex        *int      `json:"selected_index"`
	IsCorrect            *bool     `json:"is_correct"`
	MarksAwarded         float64   `gorm:"default:0" json:"marks_awarded"`
}

// CodingSubmission is a code answer plus the judge verdict. submission_id is NULL for practice runs.
type CodingSubmission struct {
	Base
	SubmissionID         *uuid.UUID `gorm:"type:uuid;index" json:"submission_id"`
	AssessmentQuestionID *uuid.UUID `gorm:"type:uuid" json:"assessment_question_id"`
	QuestionID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"question_id"`
	StudentID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"student_id"`
	Language             string     `gorm:"size:20;not null" json:"language"`
	SourceCode           string     `gorm:"type:text;not null" json:"source_code"`
	Status               string     `gorm:"size:30;default:'pending'" json:"status"`
	PassedCases          int        `gorm:"default:0" json:"passed_cases"`
	TotalCases           int        `gorm:"default:0" json:"total_cases"`
	MarksAwarded         float64    `gorm:"default:0" json:"marks_awarded"`
	RuntimeMS            int        `json:"runtime_ms"`
	MemoryKB             int        `json:"memory_kb"`
	Verdict              JSON       `gorm:"type:jsonb" json:"verdict"`
	IsPractice           bool       `gorm:"default:false" json:"is_practice"`
	// AttemptCount tracks how many times the student submitted code for this question.
	AttemptCount   int `gorm:"default:1" json:"attempt_count"`
	// FailedAttempts counts prior submits that did not achieve full acceptance,
	// used by the "attempt_penalty" scoring mode to apply a 10%/attempt deduction.
	FailedAttempts int `gorm:"default:0" json:"failed_attempts"`
}

// AssessmentResult is the final, rankable score per student per assessment.
type AssessmentResult struct {
	Base
	AssessmentID uuid.UUID `gorm:"type:uuid;not null;index:idx_res_unique,unique" json:"assessment_id"`
	StudentID    uuid.UUID `gorm:"type:uuid;not null;index:idx_res_unique,unique" json:"student_id"`
	Student      *Student  `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	MarksScored  float64   `gorm:"default:0" json:"marks_scored"`
	TotalMarks   int       `gorm:"default:0" json:"total_marks"`
	Percentage   float64   `gorm:"default:0" json:"percentage"`
	Rank         int       `json:"rank"`
	CorrectCount int       `gorm:"default:0" json:"correct_count"`
	WrongCount   int       `gorm:"default:0" json:"wrong_count"`
	PassedCases  int       `gorm:"default:0" json:"passed_cases"`
	Passed       bool      `gorm:"default:false" json:"passed"`
	Published    bool      `gorm:"default:false" json:"published"`
}
