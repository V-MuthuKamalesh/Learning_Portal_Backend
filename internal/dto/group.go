package dto

// CreateGroupRequest creates a student cohort.
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// UpdateGroupRequest edits a group.
type UpdateGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
}

// AddGroupMembersRequest bulk-adds students to a group.
type AddGroupMembersRequest struct {
	StudentIDs []string `json:"student_ids" binding:"required,min=1"`
}
