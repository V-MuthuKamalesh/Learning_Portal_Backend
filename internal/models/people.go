package models

import (
	"time"

	"github.com/google/uuid"
)

// Admin is a back-office user with a role and (optionally) a department scope.
type Admin struct {
	Base
	CollegeID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"college_id"`
	RoleID         uuid.UUID    `gorm:"type:uuid;not null" json:"role_id"`
	Role           *Role        `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Name           string       `gorm:"size:150;not null" json:"name"`
	Email          string       `gorm:"size:200;not null;index:idx_admin_college_email,unique" json:"email"`
	PasswordHash   string       `gorm:"size:255;not null" json:"-"`
	Phone          string       `gorm:"size:40" json:"phone"`
	IsActive       bool         `gorm:"default:true" json:"is_active"`
	EmailVerified  bool         `gorm:"default:false" json:"email_verified"`
	FailedAttempts int          `gorm:"default:0" json:"-"`
	LockedUntil    *time.Time   `json:"-"`
	LastLoginAt    *time.Time   `json:"last_login_at"`
	Departments    []Department `gorm:"many2many:admin_departments;" json:"departments,omitempty"`
}

// Student is the primary end user who takes assessments.
type Student struct {
	Base
	CollegeID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"college_id"`
	DepartmentID   *uuid.UUID  `gorm:"type:uuid;index" json:"department_id"`
	Department     *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	BatchID        *uuid.UUID  `gorm:"type:uuid;index" json:"batch_id"`
	Batch          *Batch      `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	Name           string      `gorm:"size:150;not null" json:"name"`
	RegisterNumber string      `gorm:"size:60;not null;index:idx_stu_college_reg,unique" json:"register_number"`
	Email          string      `gorm:"size:200;not null;index:idx_stu_college_email,unique" json:"email"`
	PasswordHash   string      `gorm:"size:255;not null" json:"-"`
	Year           string      `gorm:"size:20" json:"year"`
	Section        string      `gorm:"size:10" json:"section"`
	Phone          string      `gorm:"size:40" json:"phone"`
	IsActive       bool        `gorm:"default:true" json:"is_active"`
	EmailVerified  bool        `gorm:"default:false" json:"email_verified"`
	FailedAttempts int         `gorm:"default:0" json:"-"`
	LockedUntil    *time.Time  `json:"-"`
	LastLoginAt    *time.Time  `json:"last_login_at"`
	Groups         []Group     `gorm:"many2many:student_groups;" json:"groups,omitempty"`
}

// Group is an arbitrary cohort (class, batch, placement, custom).
type Group struct {
	Base
	CollegeID   uuid.UUID `gorm:"type:uuid;not null;index" json:"college_id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Description string    `json:"description"`
	Type        string    `gorm:"size:40" json:"type"`
	MemberCount int       `gorm:"-" json:"member_count"` // populated on read
}

// StudentGroup is the join row (kept explicit for created_at + uniqueness).
type StudentGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	StudentID uuid.UUID `gorm:"type:uuid;not null;index:idx_sg_unique,unique" json:"student_id"`
	GroupID   uuid.UUID `gorm:"type:uuid;not null;index:idx_sg_unique,unique" json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
}
