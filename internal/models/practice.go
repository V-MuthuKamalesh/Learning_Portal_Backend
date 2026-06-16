package models

import (
	"time"

	"github.com/google/uuid"
)

// PracticeModule is a topic-based collection of practice questions (Arrays, Trees, ...).
type PracticeModule struct {
	Base
	CollegeID   uuid.UUID `gorm:"type:uuid;not null;index" json:"college_id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Category    string    `gorm:"size:40;not null;default:'coding'" json:"category"` // coding | mcq
	Description string    `json:"description"`
	Ord         int       `gorm:"default:0" json:"ord"`
	Questions   []Question `gorm:"many2many:module_questions;" json:"questions,omitempty"`
}

// ModuleQuestion is the ordered join between modules and questions.
type ModuleQuestion struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ModuleID   uuid.UUID `gorm:"type:uuid;not null;index:idx_mq_unique,unique" json:"module_id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index:idx_mq_unique,unique" json:"question_id"`
	Ord        int       `gorm:"default:0" json:"ord"`
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
