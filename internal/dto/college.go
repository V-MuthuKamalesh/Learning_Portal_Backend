package dto

// CollegeBranding is the public payload that themes the student login page.
type CollegeBranding struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	LogoURL        string `json:"logo_url"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
}

// CreateCollegeRequest creates a tenant (super admin only).
type CreateCollegeRequest struct {
	Name           string `json:"name" binding:"required"`
	Code           string `json:"code" binding:"required"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	ContactEmail   string `json:"contact_email" binding:"omitempty,email"`
	ContactPhone   string `json:"contact_phone"`
	Address        string `json:"address"`
}

// UpdateCollegeRequest updates tenant settings/theme.
type UpdateCollegeRequest struct {
	Name           *string `json:"name"`
	PrimaryColor   *string `json:"primary_color"`
	SecondaryColor *string `json:"secondary_color"`
	ContactEmail   *string `json:"contact_email" binding:"omitempty,email"`
	ContactPhone   *string `json:"contact_phone"`
	Address        *string `json:"address"`
	IsActive       *bool   `json:"is_active"`
}

// DepartmentRequest is the create/update body for departments.
type DepartmentRequest struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

// BatchRequest is the create/update body for batches.
type BatchRequest struct {
	Name      string `json:"name" binding:"required"`
	StartYear int    `json:"start_year"`
	EndYear   int    `json:"end_year"`
}
