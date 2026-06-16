package models

import "github.com/google/uuid"

// College is the tenant root. Every other entity is scoped to a college.
type College struct {
	Base
	Name           string `gorm:"size:200;not null" json:"name"`
	Code           string `gorm:"size:50;not null;uniqueIndex" json:"code"`
	LogoURL        string `gorm:"size:500" json:"logo_url"`
	PrimaryColor   string `gorm:"size:16;default:'#4f46e5'" json:"primary_color"`
	SecondaryColor string `gorm:"size:16;default:'#0ea5e9'" json:"secondary_color"`
	ContactEmail   string `gorm:"size:200" json:"contact_email"`
	ContactPhone   string `gorm:"size:40" json:"contact_phone"`
	Address        string `json:"address"`
	IsActive       bool   `gorm:"default:true" json:"is_active"`
}

// Department within a college.
type Department struct {
	Base
	CollegeID uuid.UUID `gorm:"type:uuid;not null;index" json:"college_id"`
	Name      string    `gorm:"size:150;not null" json:"name"`
	Code      string    `gorm:"size:40;not null" json:"code"`
}

// Batch groups students by admission/graduation year.
type Batch struct {
	Base
	CollegeID uuid.UUID `gorm:"type:uuid;not null;index" json:"college_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	StartYear int       `json:"start_year"`
	EndYear   int       `json:"end_year"`
}
