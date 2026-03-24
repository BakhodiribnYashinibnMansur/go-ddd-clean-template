package domain

import (
	"context"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// RoleRepository is the write-side repository for the Role aggregate.
type RoleRepository interface {
	Save(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*Role, int64, error)
}

// PermissionRepository is the write-side repository for Permission entities.
type PermissionRepository interface {
	Save(ctx context.Context, perm *Permission) error
	FindByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	Update(ctx context.Context, perm *Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*Permission, int64, error)
}

// PolicyRepository is the write-side repository for Policy entities.
type PolicyRepository interface {
	Save(ctx context.Context, policy *Policy) error
	FindByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	Update(ctx context.Context, policy *Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*Policy, int64, error)
	FindByPermissionID(ctx context.Context, permissionID uuid.UUID) ([]*Policy, error)
}

// ScopeRepository is the write-side repository for Scope value objects.
type ScopeRepository interface {
	Save(ctx context.Context, scope Scope) error
	Delete(ctx context.Context, path, method string) error
	List(ctx context.Context, pagination shared.Pagination) ([]Scope, int64, error)
}

// RolePermissionRepository manages role-permission assignments.
type RolePermissionRepository interface {
	Assign(ctx context.Context, roleID, permissionID uuid.UUID) error
	Revoke(ctx context.Context, roleID, permissionID uuid.UUID) error
}

// PermissionScopeRepository manages permission-scope assignments.
type PermissionScopeRepository interface {
	Assign(ctx context.Context, permissionID uuid.UUID, path, method string) error
	Revoke(ctx context.Context, permissionID uuid.UUID, path, method string) error
}

// --- Read-side views ---

// RoleView is a read-model DTO for roles.
type RoleView struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

// PermissionView is a read-model DTO for permissions.
type PermissionView struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
}

// PolicyView is a read-model DTO for policies.
type PolicyView struct {
	ID           uuid.UUID      `json:"id"`
	PermissionID uuid.UUID      `json:"permission_id"`
	Effect       string         `json:"effect"`
	Priority     int            `json:"priority"`
	Active       bool           `json:"active"`
	Conditions   map[string]any `json:"conditions,omitempty"`
}

// ScopeView is a read-model DTO for scopes.
type ScopeView struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// AuthzReadRepository is the read-side repository for the authz bounded context.
type AuthzReadRepository interface {
	GetRole(ctx context.Context, id uuid.UUID) (*RoleView, error)
	ListRoles(ctx context.Context, pagination shared.Pagination) ([]*RoleView, int64, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*PermissionView, error)
	ListPermissions(ctx context.Context, pagination shared.Pagination) ([]*PermissionView, int64, error)
	ListPolicies(ctx context.Context, pagination shared.Pagination) ([]*PolicyView, int64, error)
	ListScopes(ctx context.Context, pagination shared.Pagination) ([]*ScopeView, int64, error)
}
