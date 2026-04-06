package http

import (
	authzentity "gct/internal/context/iam/generic/authz/domain/entity"
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
	Name        string               `json:"name" binding:"required"`
	ParentID    *authzentity.PermissionID `json:"parent_id,omitempty"`
	Description *string              `json:"description,omitempty"`
}

// CreatePolicyRequest is the request DTO for creating a policy.
type CreatePolicyRequest struct {
	PermissionID authzentity.PermissionID `json:"permission_id" binding:"required"`
	Effect       authzentity.PolicyEffect `json:"effect" binding:"required"`
	Priority     int                 `json:"priority"`
	Conditions   map[string]any      `json:"conditions,omitempty"`
}

// UpdatePolicyRequest is the request DTO for updating a policy.
type UpdatePolicyRequest struct {
	Effect     *authzentity.PolicyEffect `json:"effect,omitempty"`
	Priority   *int                 `json:"priority,omitempty"`
	Conditions map[string]any       `json:"conditions,omitempty"`
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
	PermissionID authzentity.PermissionID `json:"permission_id" binding:"required"`
}

// AssignScopeRequest is the request DTO for assigning a scope to a permission.
type AssignScopeRequest struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}
