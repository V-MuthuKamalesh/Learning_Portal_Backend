package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification types.
const (
	NotifAssessmentReminder = "assessment_reminder"
	NotifAssessmentDone     = "assessment_completed"
	NotifResultPublished    = "result_published"
	NotifGeneric            = "generic"
)

// Notification is an in-app message for a student or admin.
type Notification struct {
	Base
	CollegeID uuid.UUID `gorm:"type:uuid;not null;index" json:"college_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_notif_user" json:"user_id"`
	UserType  string    `gorm:"size:10;not null;index:idx_notif_user" json:"user_type"`
	Type      string    `gorm:"size:40;not null" json:"type"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Body      string    `json:"body"`
	Link      string    `gorm:"size:300" json:"link"`
	IsRead    bool      `gorm:"default:false;index:idx_notif_user" json:"is_read"`
}

// ActivityLog is an immutable audit record.
type ActivityLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CollegeID  *uuid.UUID `gorm:"type:uuid;index:idx_act_college" json:"college_id"`
	ActorID    *uuid.UUID `gorm:"type:uuid" json:"actor_id"`
	ActorType  string     `gorm:"size:10" json:"actor_type"`
	Action     string     `gorm:"size:80;not null" json:"action"`
	Resource   string     `gorm:"size:80" json:"resource"`
	ResourceID *uuid.UUID `gorm:"type:uuid" json:"resource_id"`
	Metadata   JSON       `gorm:"type:jsonb" json:"metadata"`
	IP         string     `gorm:"size:64" json:"ip"`
	UserAgent  string     `gorm:"size:300" json:"user_agent"`
	CreatedAt  time.Time  `gorm:"index:idx_act_college" json:"created_at"`
}

// RefreshToken stores the SHA-256 of an opaque refresh token for rotation/revocation.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	UserType  string    `gorm:"size:10;not null" json:"user_type"`
	TokenHash string    `gorm:"size:255;not null;uniqueIndex" json:"-"`
	FamilyID  uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Revoked   bool      `gorm:"default:false" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
