package dto

import "time"

// CreateAssessmentRequest creates a draft assessment shell.
type CreateAssessmentRequest struct {
	Title                string  `json:"title" binding:"required"`
	Description          string  `json:"description"`
	Type                 string  `json:"type" binding:"required"`
	DurationMinutes      int     `json:"duration_minutes"`
	TotalMarks           int     `json:"total_marks"`
	PassingMarks         int     `json:"passing_marks"`
	NegativeMarking      bool    `json:"negative_marking"`
	NegativeMarks        float64 `json:"negative_marks"`
	ShuffleQuestions     bool    `json:"shuffle_questions"`
	CodingScoringMode    string  `json:"coding_scoring_mode"`
	MCQDurationMinutes   int     `json:"mcq_duration_minutes"`
	AllowPrevious        bool    `json:"allow_previous"`
	CodingTimingMode     string  `json:"coding_timing_mode"`
	Company              string  `json:"company"` // target company e.g. "TCS", "Infosys"
	Tags                 string  `json:"tags"`    // comma-separated: "aptitude,verbal,placement"
}

// UpdateAssessmentRequest edits assessment metadata and schedule.
type UpdateAssessmentRequest struct {
	Title                *string    `json:"title"`
	Description          *string    `json:"description"`
	Type                 *string    `json:"type"`
	DurationMinutes      *int       `json:"duration_minutes"`
	TotalMarks           *int       `json:"total_marks"`
	PassingMarks         *int       `json:"passing_marks"`
	NegativeMarking      *bool      `json:"negative_marking"`
	NegativeMarks        *float64   `json:"negative_marks"`
	ShuffleQuestions     *bool      `json:"shuffle_questions"`
	CodingScoringMode    *string    `json:"coding_scoring_mode"`
	MCQDurationMinutes   *int       `json:"mcq_duration_minutes"`
	AllowPrevious        *bool      `json:"allow_previous"`
	CodingTimingMode     *string    `json:"coding_timing_mode"`
	StartTime            *time.Time `json:"start_time"`
	EndTime              *time.Time `json:"end_time"`
	Company              *string    `json:"company"`
	Tags                 *string    `json:"tags"`
}

// UpdateAssessmentQuestionRequest patches per-question slot settings.
type UpdateAssessmentQuestionRequest struct {
	CodingTimeLimitMinutes *int `json:"coding_time_limit_minutes"`
	Marks                  *int `json:"marks"`
	Ord                    *int `json:"ord"`
}

// AttachQuestionsRequest links questions to an assessment.
type AttachQuestionsRequest struct {
	QuestionIDs []string `json:"question_ids" binding:"required,min=1"`
}

// AssignAssessmentRequest targets cohorts.
type AssignAssessmentRequest struct {
	TargetType string  `json:"target_type" binding:"required"` // college|student|department|group|batch
	TargetID   *string `json:"target_id"`
}
