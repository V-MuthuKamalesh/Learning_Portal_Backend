package dto

// CreateStudentRequest adds a single student. Initial password defaults to the
// register number unless provided (student is prompted to change on first login).
type CreateStudentRequest struct {
	Name           string  `json:"name" binding:"required"`
	RegisterNumber string  `json:"register_number" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	DepartmentID   *string `json:"department_id"`
	BatchID        *string `json:"batch_id"`
	Year           string  `json:"year"`
	Section        string  `json:"section"`
	Phone          string  `json:"phone"`
	Password       string  `json:"password"`
}

// UpdateStudentRequest edits a student.
type UpdateStudentRequest struct {
	Name         *string `json:"name"`
	Email        *string `json:"email" binding:"omitempty,email"`
	DepartmentID *string `json:"department_id"`
	BatchID      *string `json:"batch_id"`
	Year         *string `json:"year"`
	Section      *string `json:"section"`
	Phone        *string `json:"phone"`
}

// StudentStatusRequest activates/deactivates an account.
type StudentStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// StudentFilter captures list query filters.
type StudentFilter struct {
	DepartmentID string
	BatchID      string
	GroupID      string
	Year         string
	Section      string
	Status       string // active | inactive
}

// BulkImportResult reports the outcome of a CSV import.
type BulkImportResult struct {
	Created int      `json:"created"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}
