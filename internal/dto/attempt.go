package dto

import "time"

// SaveAnswerRequest autosaves an MCQ selection.
type SaveAnswerRequest struct {
	AssessmentQuestionID string `json:"assessment_question_id" binding:"required"`
	SelectedIndex        int    `json:"selected_index"`
}

// SubmitAttemptRequest finalizes an attempt.
type SubmitAttemptRequest struct {
	Auto bool `json:"auto"`
}

// StudentAssessmentView is an assessment row for the student portal.
type StudentAssessmentView struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	Workflow        string     `json:"workflow"`
	DurationMinutes int        `json:"duration_minutes"`
	TotalMarks      int        `json:"total_marks"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	Company         string     `json:"company,omitempty"`
	Tags            string     `json:"tags,omitempty"`
	SubmissionID    *string    `json:"submission_id,omitempty"`
	AttemptStatus   string     `json:"attempt_status,omitempty"`
	Score           *float64   `json:"score,omitempty"`
}

// AttemptDetail is the in-progress attempt payload.
type AttemptDetail struct {
	ID                   string                          `json:"id"`
	AssessmentID         string                          `json:"assessment_id"`
	Status               string                          `json:"status"`
	StartedAt            time.Time                       `json:"started_at"`
	ExpiresAt            time.Time                       `json:"expires_at"`
	Questions            []AttemptQuestion               `json:"questions"`
	Answers              map[string]int                  `json:"answers"`
	Coding               map[string]CodingSubmissionView `json:"coding,omitempty"`
	CodingScoringMode    string                          `json:"coding_scoring_mode,omitempty"`
	MCQDurationMinutes   int                             `json:"mcq_duration_minutes"`
	AllowPrevious        bool                            `json:"allow_previous"`
	CodingTimingMode     string                          `json:"coding_timing_mode"`
}

// AttemptQuestion is a sanitized question for students.
type AttemptQuestion struct {
	ID                   string   `json:"id"`
	AssessmentQuestionID string   `json:"assessment_question_id"`
	Type                 string   `json:"type"`
	Body                 string   `json:"body,omitempty"`
	Options              []string `json:"options,omitempty"`
	Title                string   `json:"title,omitempty"`
	Description          string   `json:"description,omitempty"`
	InputFormat          string   `json:"input_format,omitempty"`
	OutputFormat         string   `json:"output_format,omitempty"`
	Constraints          string   `json:"constraints,omitempty"`
	SampleInput          string   `json:"sample_input,omitempty"`
	SampleOutput         string   `json:"sample_output,omitempty"`
	TimeLimitMS              int `json:"time_limit_ms,omitempty"`
	MemoryLimitMB            int `json:"memory_limit_mb,omitempty"`
	CodingTimeLimitMinutes   int `json:"coding_time_limit_minutes,omitempty"`
	Marks                    int `json:"marks"`
	Ord                      int `json:"ord"`
}

// CodingSubmissionView is the student's saved code for a question.
type CodingSubmissionView struct {
	Language       string  `json:"language"`
	SourceCode     string  `json:"source_code"`
	Status         string  `json:"status"`
	PassedCases    int     `json:"passed_cases"`
	TotalCases     int     `json:"total_cases"`
	MarksAwarded   float64 `json:"marks_awarded"`
	AttemptCount   int     `json:"attempt_count"`
	FailedAttempts int     `json:"failed_attempts"`
}

// TestCaseResult is one judge verdict row.
type TestCaseResult struct {
	Ord            int    `json:"ord"`
	Passed         bool   `json:"passed"`
	IsHidden       bool   `json:"is_hidden"`
	Input          string `json:"input,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	ActualOutput   string `json:"actual_output,omitempty"`
	Status         string `json:"status"`
	Weight         int    `json:"weight"`
}

// CodingRunResult is returned from run/submit endpoints.
type CodingRunResult struct {
	Status         string           `json:"status"`
	PassedCases    int              `json:"passed_cases"`
	TotalCases     int              `json:"total_cases"`
	MarksAwarded   float64          `json:"marks_awarded"`
	Results        []TestCaseResult `json:"results"`
	Stdout         string           `json:"stdout,omitempty"`
	Stderr         string           `json:"stderr,omitempty"`
	RuntimeMS      int              `json:"runtime_ms,omitempty"`
	AttemptCount   int              `json:"attempt_count,omitempty"`
	FailedAttempts int              `json:"failed_attempts,omitempty"`
}
