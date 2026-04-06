package repository

import (
	"context"

	"gct/internal/context/iam/generic/authz/domain/entity"
	shared "gct/internal/kernel/domain"
)

// --- Read-side views ---

// RoleView is a read-model DTO for roles.
type RoleView struct {
	ID          entity.RoleID `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
}

// PermissionView is a read-model DTO for permissions.
type PermissionView struct {
	ID          entity.PermissionID  `json:"id"`
	ParentID    *entity.PermissionID `json:"parent_id,omitempty"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
}

// PolicyView is a read-model DTO for policies.
type PolicyView struct {
	ID           entity.PolicyID     `json:"id"`
	PermissionID entity.PermissionID `json:"permission_id"`
	Effect       string              `json:"effect"`
	Priority     int                 `json:"priority"`
	Active       bool                `json:"active"`
	Conditions   map[string]any      `json:"conditions,omitempty"`
}

// ScopeView is a read-model DTO for scopes.
type ScopeView struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// AuthzReadRepository is the read-side (CQRS query) repository for the entire authz bounded context.
// It consolidates role, permission, policy, and scope reads into a single interface to simplify
// dependency injection for query handlers.
type AuthzReadRepository interface {
	GetRole(ctx context.Context, id entity.RoleID) (*RoleView, error)
	ListRoles(ctx context.Context, pagination shared.Pagination) ([]*RoleView, int64, error)
	GetPermission(ctx context.Context, id entity.PermissionID) (*PermissionView, error)
	ListPermissions(ctx context.Context, pagination shared.Pagination) ([]*PermissionView, int64, error)
	ListPolicies(ctx context.Context, pagination shared.Pagination) ([]*PolicyView, int64, error)
	ListScopes(ctx context.Context, pagination shared.Pagination) ([]*ScopeView, int64, error)
	CheckAccess(ctx context.Context, roleID entity.RoleID, path, method string, evalCtx entity.EvaluationContext) (bool, error)
	FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []entity.PermissionID) ([]*entity.Policy, error)
}
