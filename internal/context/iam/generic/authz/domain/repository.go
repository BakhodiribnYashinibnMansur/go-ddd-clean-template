package domain

import (
	"context"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// RoleRepository is the write-side repository for the Role aggregate.
// Implementations must return ErrRoleNotFound from FindByID when no row matches.
// Save persists a new role; Update persists changes to an existing one including its child permissions.
type RoleRepository interface {
	Save(ctx context.Context, role *Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*Role, int64, error)
}

// PermissionRepository is the write-side repository for Permission entities.
// Permissions may exist independently of roles (referenced via join tables), so this
// repository manages their full lifecycle separate from the Role aggregate.
type PermissionRepository interface {
	Save(ctx context.Context, perm *Permission) error
	FindByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	Update(ctx context.Context, perm *Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*Permission, int64, error)
}

// PolicyRepository is the write-side repository for ABAC Policy entities.
// FindByPermissionID returns all policies bound to a given permission, enabling bulk evaluation.
type PolicyRepository interface {
	Save(ctx context.Context, policy *Policy) error
	FindByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	Update(ctx context.Context, policy *Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*Policy, int64, error)
	FindByPermissionID(ctx context.Context, permissionID uuid.UUID) ([]*Policy, error)
}

// ScopeRepository persists Scope value objects as a global registry of API endpoints.
// Scopes are identified by their composite key (path + method), not by a UUID.
type ScopeRepository interface {
	Save(ctx context.Context, scope Scope) error
	Delete(ctx context.Context, path, method string) error
	List(ctx context.Context, pagination shared.Pagination) ([]Scope, int64, error)
}

// RolePermissionRepository manages the many-to-many join between roles and permissions.
// Assign and Revoke are idempotent — assigning an already-assigned pair is a no-op.
type RolePermissionRepository interface {
	Assign(ctx context.Context, roleID, permissionID uuid.UUID) error
	Revoke(ctx context.Context, roleID, permissionID uuid.UUID) error
}

// PermissionScopeRepository manages the many-to-many join between permissions and scopes.
// The composite key is (permissionID, path, method).
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

// AuthzReadRepository is the read-side (CQRS query) repository for the entire authz bounded context.
// It consolidates role, permission, policy, and scope reads into a single interface to simplify
// dependency injection for query handlers.
type AuthzReadRepository interface {
	GetRole(ctx context.Context, id uuid.UUID) (*RoleView, error)
	ListRoles(ctx context.Context, pagination shared.Pagination) ([]*RoleView, int64, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*PermissionView, error)
	ListPermissions(ctx context.Context, pagination shared.Pagination) ([]*PermissionView, int64, error)
	ListPolicies(ctx context.Context, pagination shared.Pagination) ([]*PolicyView, int64, error)
	ListScopes(ctx context.Context, pagination shared.Pagination) ([]*ScopeView, int64, error)
	CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx EvaluationContext) (bool, error)
	FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []uuid.UUID) ([]*Policy, error)
}
