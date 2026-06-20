package models

import (
	"time"

	"github.com/google/uuid"
)

// Assessment type & status enums.
const (
	AssessmentMCQ         = "mcq"
	AssessmentProgramming = "programming"
	AssessmentMixed       = "mixed"

	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusScheduled = "scheduled"

	TargetStudent    = "student"
	TargetDepartment = "department"
	TargetGroup      = "group"
	TargetBatch      = "batch"
	TargetCollege    = "college"
)

// Assessment is a scheduled test composed of questions and assigned to cohorts.
type Assessment struct {
	Base
	CollegeID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"college_id"`
	Title            string     `gorm:"size:200;not null" json:"title"`
	Description      string     `json:"description"`
	Type             string     `gorm:"size:20;not null" json:"type"`
	StartTime        *time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	DurationMinutes  int        `gorm:"default:60" json:"duration_minutes"`
	TotalMarks       int        `gorm:"default:0" json:"total_marks"`
	PassingMarks     int        `gorm:"default:0" json:"passing_marks"`
	MCQCount         int        `gorm:"default:0" json:"mcq_count"`
	CodingCount      int        `gorm:"default:0" json:"coding_count"`
	ShuffleQuestions  bool       `gorm:"default:false" json:"shuffle_questions"`
	NegativeMarking   bool       `gorm:"default:false" json:"negative_marking"`
	NegativeMarks     float64    `gorm:"default:0" json:"negative_marks"`
	AutoSubmit        bool       `gorm:"default:true" json:"auto_submit"`
	// CodingScoringMode controls how coding question marks are calculated.
	// "weighted" = marks proportional to test-case weights (default)
	// "attempt_penalty" = 10% mark deduction per prior failed full submission
	CodingScoringMode string     `gorm:"size:30;default:'weighted'" json:"coding_scoring_mode"`
	Status            string     `gorm:"size:20;not null;default:'draft'" json:"status"`
	CreatedBy        *uuid.UUID `gorm:"type:uuid" json:"created_by"`

	Questions   []AssessmentQuestion   `gorm:"foreignKey:AssessmentID" json:"questions,omitempty"`
	Assignments []AssessmentAssignment `gorm:"foreignKey:AssessmentID" json:"assignments,omitempty"`
}

// Workflow returns the runtime lifecycle state relative to now.
// upcoming | running | completed | expired — derived from times + status.
func (a *Assessment) Workflow(now time.Time) string {
	if a.Status == StatusDraft {
		return "draft"
	}
	if a.StartTime != nil && now.Before(*a.StartTime) {
		return "upcoming"
	}
	if a.EndTime != nil && now.After(*a.EndTime) {
		return "expired"
	}
	if a.StartTime != nil && a.EndTime != nil && !now.Before(*a.StartTime) && !now.After(*a.EndTime) {
		return "running"
	}
	return "scheduled"
}

// AssessmentQuestion links a question into an assessment with order & mark override.
type AssessmentQuestion struct {
	Base
	AssessmentID uuid.UUID `gorm:"type:uuid;not null;index:idx_aq_unique,unique" json:"assessment_id"`
	QuestionID   uuid.UUID `gorm:"type:uuid;not null;index:idx_aq_unique,unique" json:"question_id"`
	Question     *Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Ord          int       `gorm:"default:0" json:"ord"`
	Marks        *int      `json:"marks"`
}

// AssessmentAssignment targets an assessment at a cohort.
type AssessmentAssignment struct {
	Base
	AssessmentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"assessment_id"`
	TargetType   string     `gorm:"size:20;not null" json:"target_type"`
	TargetID     *uuid.UUID `gorm:"type:uuid" json:"target_id"`
}
