package dto

import "time"

// CreateAssessmentRequest creates a draft assessment shell.
type CreateAssessmentRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	Type            string `json:"type" binding:"required"`
	DurationMinutes int    `json:"duration_minutes"`
	TotalMarks      int    `json:"total_marks"`
}

// UpdateAssessmentRequest edits assessment metadata and schedule.
type UpdateAssessmentRequest struct {
	Title           *string    `json:"title"`
	Description     *string    `json:"description"`
	Type            *string    `json:"type"`
	DurationMinutes *int       `json:"duration_minutes"`
	TotalMarks      *int       `json:"total_marks"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
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
