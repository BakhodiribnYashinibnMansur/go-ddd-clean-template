package http

import (
	"gct/internal/authz/domain"

	"github.com/google/uuid"
)

// CreateRoleRequest is the request DTO for creating a role.
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

// UpdateRoleRequest is the request DTO for updating a role.
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreatePermissionRequest is the request DTO for creating a permission.
type CreatePermissionRequest struct {
	Name        string     `json:"name" binding:"required"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Description *string    `json:"description,omitempty"`
}

// CreatePolicyRequest is the request DTO for creating a policy.
type CreatePolicyRequest struct {
	PermissionID uuid.UUID          `json:"permission_id" binding:"required"`
	Effect       domain.PolicyEffect `json:"effect" binding:"required"`
	Priority     int                `json:"priority"`
	Conditions   map[string]string   `json:"conditions,omitempty"`
}

// UpdatePolicyRequest is the request DTO for updating a policy.
type UpdatePolicyRequest struct {
	Effect     *domain.PolicyEffect `json:"effect,omitempty"`
	Priority   *int                 `json:"priority,omitempty"`
	Conditions map[string]string     `json:"conditions,omitempty"`
}

// CreateScopeRequest is the request DTO for creating a scope.
type CreateScopeRequest struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

// DeleteScopeRequest is the request DTO for deleting a scope.
type DeleteScopeRequest struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

// AssignPermissionRequest is the request DTO for assigning a permission to a role.
type AssignPermissionRequest struct {
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
}

// AssignScopeRequest is the request DTO for assigning a scope to a permission.
type AssignScopeRequest struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}
