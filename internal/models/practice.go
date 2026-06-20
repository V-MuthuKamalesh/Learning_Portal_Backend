package models

import (
	"time"

	"github.com/google/uuid"
)

// PracticeModule is a topic-based collection of practice questions (Arrays, Trees, ...).
type PracticeModule struct {
	Base
	CollegeID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"college_id"`
	Name        string     `gorm:"size:120;not null" json:"name"`
	Category    string     `gorm:"size:40;not null;default:'mixed'" json:"category"` // coding | mcq | mixed
	Description string     `json:"description"`
	Tags        string     `gorm:"size:255" json:"tags"`        // comma-separated topic tags
	IsPublished bool       `gorm:"default:false" json:"is_published"`
	Ord         int        `gorm:"default:0" json:"ord"`
	Questions   []Question `gorm:"many2many:module_questions;" json:"questions,omitempty"`
}

// ModuleQuestion is the ordered join between modules and questions,
// with per-slot overrides for marks and attempt limits.
type ModuleQuestion struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ModuleID    uuid.UUID `gorm:"type:uuid;not null;index:idx_mq_unique,unique" json:"module_id"`
	QuestionID  uuid.UUID `gorm:"type:uuid;not null;index:idx_mq_unique,unique" json:"question_id"`
	Ord         int       `gorm:"default:0" json:"ord"`
	Marks       int       `gorm:"default:0" json:"marks"`        // 0 = inherit from question
	MaxAttempts int       `gorm:"default:0" json:"max_attempts"` // 0 = unlimited
}

// StudentProgress tracks completion of a practice module per student.
type StudentProgress struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	StudentID  uuid.UUID `gorm:"type:uuid;not null;index:idx_sp_unique,unique" json:"student_id"`
	ModuleID   uuid.UUID `gorm:"type:uuid;not null;index:idx_sp_unique,unique" json:"module_id"`
	Completed  int       `gorm:"default:0" json:"completed"`
	Total      int       `gorm:"default:0" json:"total"`
	Percentage float64   `gorm:"default:0" json:"percentage"`
	UpdatedAt  time.Time `json:"updated_at"`
}
