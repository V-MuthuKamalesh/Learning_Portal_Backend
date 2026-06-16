package dto

// CreateAdminRequest creates a college admin with a system role.
type CreateAdminRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RoleID   string `json:"role_id" binding:"required"`
	Phone    string `json:"phone"`
}

// UpdateAdminRequest edits admin profile fields.
type UpdateAdminRequest struct {
	Name  *string `json:"name"`
	Phone *string `json:"phone"`
}

// AdminStatusRequest activates/deactivates an admin.
type AdminStatusRequest struct {
	IsActive bool `json:"is_active"`
}
