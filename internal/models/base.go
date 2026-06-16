package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base is embedded by every entity: UUID primary key, timestamps, soft delete.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate guarantees an ID even when the DB default is unavailable.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// AllModels returns every entity for auto-migration, in dependency order.
func AllModels() []any {
	return []any{
		&College{}, &Department{}, &Batch{},
		&Permission{}, &Role{},
		&Admin{}, &Student{}, &Group{}, &StudentGroup{},
		&Question{}, &MCQQuestion{}, &ProgrammingQuestion{}, &TestCase{},
		&PracticeModule{}, &ModuleQuestion{}, &StudentProgress{},
		&Assessment{}, &AssessmentQuestion{}, &AssessmentAssignment{},
		&Submission{}, &Answer{}, &CodingSubmission{}, &AssessmentResult{},
		&Notification{}, &ActivityLog{}, &RefreshToken{},
	}
}
