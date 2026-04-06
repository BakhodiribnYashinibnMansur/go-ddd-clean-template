package dto

import (
	"github.com/google/uuid"
)

// RoleView is a read-model DTO returned by query handlers.
type RoleView struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

// PermissionView is a read-model DTO returned by query handlers.
type PermissionView struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
}

// PolicyView is a read-model DTO returned by query handlers.
type PolicyView struct {
	ID           uuid.UUID      `json:"id"`
	PermissionID uuid.UUID      `json:"permission_id"`
	Effect       string         `json:"effect"`
	Priority     int            `json:"priority"`
	Active       bool           `json:"active"`
	Conditions   map[string]any `json:"conditions,omitempty"`
}

// ScopeView is a read-model DTO returned by query handlers.
type ScopeView struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}
