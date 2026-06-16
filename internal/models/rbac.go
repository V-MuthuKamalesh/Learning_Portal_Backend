package models

import "github.com/google/uuid"

// Permission is a resource:action capability (e.g. student:create).
type Permission struct {
	Base
	Resource string `gorm:"size:50;not null" json:"resource"`
	Action   string `gorm:"size:50;not null" json:"action"`
	Slug     string `gorm:"size:101;not null;uniqueIndex" json:"slug"` // resource:action
	Label    string `gorm:"size:120" json:"label"`
}

// Role is a named bundle of permissions, scoped to a college (NULL = system template).
type Role struct {
	Base
	CollegeID   *uuid.UUID   `gorm:"type:uuid;index" json:"college_id,omitempty"`
	Name        string       `gorm:"size:80;not null" json:"name"`
	Slug        string       `gorm:"size:80;not null" json:"slug"`
	IsSystem    bool         `gorm:"default:false" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// PermissionSlugs flattens the role's permissions into slug strings for JWT claims.
func (r *Role) PermissionSlugs() []string {
	out := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		out = append(out, p.Slug)
	}
	return out
}

// HasWildcard reports whether the role carries the super "*" permission.
func (r *Role) HasWildcard() bool {
	for _, p := range r.Permissions {
		if p.Slug == "*" {
			return true
		}
	}
	return false
}
